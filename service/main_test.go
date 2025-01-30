package main_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	m "github.com/fionn/address-manager/service"
	"github.com/fionn/address-manager/service/fireblocks"

	"github.com/fionn/address-manager/fb_mock"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const databaseFile = "test.db"
const fbBaseHost = "localhost:6200"
const fbBaseURL = "http://" + fbBaseHost

func setupDatabase() (*gorm.DB, error) {
	os.Remove(databaseFile)
	db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&m.User{}, &m.Wallet{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setupMock(address string) (*sync.WaitGroup, context.CancelFunc) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, stopMock := context.WithCancel(context.Background())
	go fb_mock.RunWithCancellation(ctx, wg, address)
	return wg, stopMock
}

func TestPopulateWalletPool(t *testing.T) {
	wg, stopMock := setupMock(fbBaseHost)
	defer wg.Wait()
	defer stopMock()

	fb := fireblocks.NewFireblocksSession(fbBaseURL)
	threshold := 1

	walletChannel := make(chan m.Wallet, threshold)
	defer close(walletChannel)

	ctx, cancelWalletPool := context.WithCancel(context.Background())
	defer cancelWalletPool()

	go m.PopulateWalletPool(walletChannel, ctx, threshold, &fb)
	wallet := <-walletChannel

	if wallet.AddressBTC == "" {
		t.Error("Got zero-valued BTC address")
	}

	// Unpleasant. We must wait for the channel to repopulate, but we don't
	// have a need for this in the actual code so don't implement a wait group
	// or anything else that we could use here.
	time.Sleep(time.Second)
	if len(walletChannel) != threshold {
		t.Errorf("walletChannel not at threshold")
	}
}

func TestCreateUser(t *testing.T) {
	db, err := setupDatabase()
	defer os.Remove(databaseFile)
	if err != nil {
		t.Fatalf("Error instantiating the database: %s", err)
	}

	wg, stopMock := setupMock(fbBaseHost)
	defer wg.Wait()
	defer stopMock()

	fb := fireblocks.NewFireblocksSession(fbBaseURL)

	threshold := 1
	walletChannel := make(chan m.Wallet, threshold)
	defer close(walletChannel)

	ctx, cancelWalletPool := context.WithCancel(context.Background())
	defer cancelWalletPool()
	go m.PopulateWalletPool(walletChannel, ctx, threshold, &fb)

	user, err := m.CreateUser(db, walletChannel)
	if err != nil {
		t.Fatalf("Failed to create user: %s", err)
	}

	user_prime, err := m.GetUser(db, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %s", err)
	}

	if user.ID != user_prime.ID {
		t.Errorf("User IDs %s, %s do not match", user.ID, user_prime.ID)
	}

	if user.Wallet.ID != user_prime.Wallet.ID {
		t.Errorf("Wallet IDs %d, %d do not match", user.Wallet.ID, user_prime.Wallet.ID)
	}

	if user.Wallet.AddressBTC != user_prime.Wallet.AddressBTC {
		t.Errorf("Bitcoin addresses %s, %s do not match", user.Wallet.AddressBTC, user_prime.Wallet.AddressBTC)
	}
}
