package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/fionn/address-manager/service/fb"
)

const databaseFile = "test.db"

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

func newWallet() (*Wallet, error) {
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
func populateWalletPool(c chan<- Wallet, ctx context.Context, threshold int) {
	for {
		select {
		case <-ctx.Done():
			// TODO: drain channel into persistent storage.
			log.Println("Received cancellation")
			return
		default:
			for len(c) < threshold {
				wallet, _ := newWallet()
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

	db.AutoMigrate(&User{})

	threshold := 30
	walletChannel := make(chan Wallet, threshold)
	defer close(walletChannel)
	ctx, cancel := context.WithCancel(context.Background())

	// TODO: check persistent storage for unallocated wallets and add them to
	// the pool first.

	go populateWalletPool(walletChannel, ctx, threshold)

	// We don't use a WaitGroup because we don't actually want to wait, just let
	// the goroutine run until the process terminates to keep the pool
	// topped-up.

	user := User{}
	db.Create(&user)
	db.Model(&user).Update("Wallet", <-walletChannel)
	fmt.Printf("%+v\n", user)

	// TODO: catch signal and try to exit gracefully.
	cancel()
	// TODO: wait for wallet channel to be drained.
}
