package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/fionn/address-manager/service/fireblocks"
)

const databaseFile = "adhoc.db"
const fbBaseURL = "http://localhost:6200"

type Data struct {
	DB      *gorm.DB
	Wallets <-chan Wallet
}

type Wallet struct {
	// This is very lossy and we're probably better off keeping the general
	// structure of the Fireblocks API responses, but not 1:1 since we don't
	// want to rely on Fireblocks keeping their API stable for our database
	// schema.
	gorm.Model
	AddressBTC string
	UserID     uuid.UUID
}

type User struct {
	ID        uuid.UUID `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Wallet    Wallet
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	uuid := uuid.New()
	tx.Statement.SetColumn("ID", uuid)
	return nil
}

// Create a wallet.
func newWallet(fb *fireblocks.Fireblocks) (*Wallet, error) {
	fbVaultAccount, err := fb.CreateVaultAccount()
	if err != nil {
		return nil, err
	}

	// By default the above call creates an Ethereum wallet for us, but (at
	// least for now) we want a Bitcoin one, so we have to create that
	// separately.
	fbVaultWallet, err := fb.CreateVaultAccountAsset(fbVaultAccount.ID, "BTC")
	if err != nil {
		return nil, err
	}

	return &Wallet{AddressBTC: fbVaultWallet.Address}, nil
}

// Keep the wallet pool populated.
func PopulateWalletPool(c chan<- Wallet, ctx context.Context, threshold int, fb *fireblocks.Fireblocks) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Cancelling wallet pool population")
			return
		default:
			for len(c) < threshold {
				wallet, err := newWallet(fb)
				if err != nil {
					log.Printf("Failed to create wallet: %s\n", err)
					time.Sleep(1 * time.Second) // TODO: exponential backoff with cap.
					continue
				}
				c <- *wallet
			}
		}
	}
}

func (d *Data) CreateUser() (*User, error) {
	user := User{Wallet: <-d.Wallets}
	if tx := d.DB.Create(&user); tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func (d Data) GetUser(id uuid.UUID) (*User, error) {
	user := User{}
	if tx := d.DB.Model(&user).Preload("Wallet").Take(&user, id); tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

// Helper to write error messages as HTTP responses.
func writeError(w http.ResponseWriter, httpErrorCode int, message string) {
	w.WriteHeader(httpErrorCode)
	w.Write([]byte(message))
}

func (d *Data) handlePostCreateUser(w http.ResponseWriter, _ *http.Request) {
	user, err := d.CreateUser()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = w.Write(response)
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}
}

func (d Data) handleGetUser(w http.ResponseWriter, r *http.Request) {
	userId, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := d.GetUser(userId)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	response, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = w.Write(response)
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}
}

// Entry point (still for testing).
func main() {
	os.Remove(databaseFile)
	db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
	defer os.Remove(databaseFile)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %s", err)
	}

	fb := fireblocks.NewFireblocksSession(fbBaseURL)

	err = db.AutoMigrate(&User{}, &Wallet{})
	if err != nil {
		log.Fatalf("Failed to automigrate: %s", err)
	}

	threshold := 30
	walletChannel := make(chan Wallet, threshold)
	defer close(walletChannel)

	ctx, cancelWalletPool := context.WithCancel(context.Background())
	defer cancelWalletPool()
	go PopulateWalletPool(walletChannel, ctx, threshold, &fb)

	data := Data{DB: db, Wallets: walletChannel}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/user/{userId}", data.handleGetUser)
	r.Post("/user", data.handlePostCreateUser)

	address := "localhost:6201"
	log.Printf("listening on http://%s/", address)
	if err := http.ListenAndServe(address, r); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
