package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/fionn/address-manager/service/fireblocks"
)

const databaseFile = "test.db"
const FBBaseURL = "http://localhost:6200"

type Wallet struct {
	// This is very lossy and we're probably better off keeping the general
	// structure of the Fireblocks API responses, but not 1:1 since we don't
	// want to rely on Fireblocks keeping their API stable for our database
	// schema.
	gorm.Model
	AddressBTC string
}

type User struct {
	gorm.Model
	WalletID uint
	Wallet   Wallet
}

// Create a wallet.
// TODO: create the wallet persistently (i.e. don't just instantiate the object
// but write it to the database too).
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

// Keep the wallet pool polulated.
func populateWalletPool(c chan<- Wallet, ctx context.Context, threshold int, fb *fireblocks.Fireblocks) {
	for {
		select {
		case <-ctx.Done():
			// TODO: reconsider providing cancellation as it shouldn't be
			// necessary if all wallets are committed to the database on
			// creation.
			log.Println("Received cancellation")
			return
		default:
			for len(c) < threshold {
				wallet, _ := newWallet(fb)
				c <- *wallet
			}
		}
	}
}

// This entry point exists for testing. We remove the on-disk database file if
// it exists and create a new one, then add some example data.
func main() {
	os.Remove(databaseFile)
	db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %s", err)
	}

	fb := fireblocks.NewFireblocksSession(FBBaseURL)

	err = db.AutoMigrate(&User{}, &Wallet{})
	if err != nil {
		log.Fatalf("Failed to automigrate: %s", err)
	}

	// TODO: check persistent storage for unallocated wallets and add them to
	// the pool first.

	threshold := 30
	walletChannel := make(chan Wallet, threshold)
	defer close(walletChannel)
	ctx, cancelWalletPool := context.WithCancel(context.Background())
	go populateWalletPool(walletChannel, ctx, threshold, &fb)

	user := User{}
	db.Create(&user)
	db.Model(&user).Update("Wallet", <-walletChannel)
	fmt.Printf("%+v\n", user)

	// TODO: catch signal and try to exit gracefully.
	cancelWalletPool()
	// TODO: wait for cancellation to complete.
}
