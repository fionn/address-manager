package main

import (
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

// This entry point exists for testing. We remove the on-disk database file if
// it exists and create a new one, then add some example data.
func main() {
	os.Remove(databaseFile)
	db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %s", err)
	}

	wallet, _ := newWallet()
	fmt.Printf("%+v\n", wallet)

	db.AutoMigrate(&User{})
}
