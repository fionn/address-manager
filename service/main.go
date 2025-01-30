package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/fionn/address-manager/service/fireblocks"
)

const databaseFile = "test.db"
const fbBaseURL = "http://localhost:6200"

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
func populateWalletPool(c chan<- Wallet, ctx context.Context, threshold int, fb *fireblocks.Fireblocks) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Received cancellation")
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

func createUser(db *gorm.DB, c <-chan Wallet) (*User, error) {
	user := User{Wallet: <-c}
	if tx := db.Create(&user); tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func getUser(db *gorm.DB, id uuid.UUID) (*User, error) {
	user := User{}
	if tx := db.Model(&user).Preload("Wallet").Take(&user, id); tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

// This entry point exists for testing. We remove the on-disk database file if
// it exists and create a new one, then add some example data.
func main() {
	os.Remove(databaseFile)
	db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
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
	go populateWalletPool(walletChannel, ctx, threshold, &fb)

	user, err := createUser(db, walletChannel)
	if err != nil {
		log.Fatalf("Failed to create user: %s", err)
	}

	fmt.Printf("%+v\n", user)

	user, err = getUser(db, user.ID)
	if err != nil {
		log.Fatalf("Failed to get user: %s", err)
	}

	fmt.Printf("%+v\n", user)
}
