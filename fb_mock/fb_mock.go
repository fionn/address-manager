package fb_mock

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	mrand "math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/btcsuite/btcutil/bech32"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	fb "github.com/fionn/address-manager/service/fireblocks"
	"github.com/fionn/address-manager/utils"
)

// Fireblocks error response.
type FBError struct {
	APIErrorCode int    `json:"error_code,omitempty"`
	Message      string `json:"message"`
}

// Generate a slice of cryptographically secure random bytes of length size.
func randomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return []byte{}, err
	}
	return b, nil
}

// Generate a random Bech32 address.
func generateBTCAddress() (string, error) {
	// Because we're not using a real keypair, this is just fed into some hash
	// functions. As such, we don't really care what it is, so just use some
	// random bytes.
	fakePubKey, err := randomBytes(20)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %s", err)
	}

	hashedPubKey := btcutil.Hash160(fakePubKey)
	witnessProgram, err := bech32.ConvertBits(hashedPubKey, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("failed to squash 8-bit array to 5-bit array: %s", err)
	}

	address, err := bech32.Encode("tb", append([]byte{0}, witnessProgram...))
	if err != nil {
		return "", fmt.Errorf("failed to encode Bech32 address: %s", err)
	}

	return address, nil
}

// Generate a random Solana address (32 bytes base-58 encoded).
func generateSOLAddress() (string, error) {
	fakePubKey, err := randomBytes(32)
	return base58.Encode(fakePubKey), err
}

func assetAddressGeneratorMap(assetId string) (func() (string, error), error) {
	if assetId == "BTC" {
		return generateBTCAddress, nil
	}
	if assetId == "SOL" {
		return generateSOLAddress, nil
	}
	return nil, errors.New("unknown asset type")
}

// Helper to write error messages as HTTP responses.
func writeError(w http.ResponseWriter, httpErrorCode int, message string, apiErrorCode int) {
	fbError, err := json.MarshalIndent(FBError{apiErrorCode, message}, "", "  ")
	if err != nil {
		err = fmt.Errorf("failed to marshal error (%d: %s): %s", apiErrorCode, message, err)
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErrorCode)
	w.Write(fbError)
}

// Handler to create a new vault. See
// https://developers.fireblocks.com/reference/createvaultaccount.
func handlePostCreateVaultAccount(w http.ResponseWriter, r *http.Request) {
	// TODO: support Idempotency-Key.

	// The documentation on this endpoint is unclear, but the example request
	// only sends two fields, both of which are explicitly optional, so we infer
	// that all fields are optional which, for our purposes, means we can ignore
	// them.
	//
	// TODO: support optional fields.

	// This isn't documented, but from the example response it seems to default
	// to creating a single ETH wallet.
	asset := fb.VaultAsset{ID: "ETH"}
	fbVaultAccount := fb.VaultAccount{ID: strconv.Itoa(mrand.Int()), Assets: []fb.VaultAsset{asset}}

	response, err := json.MarshalIndent(fbVaultAccount, "", "  ")
	if err != nil {
		log.Print(err)
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(utils.BinaryNewline(response))
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}
}

// Handler for the addresses_paginate endpoint.
// See https://developers.fireblocks.com/reference/getvaultaccountassetaddressespaginated.
func handleGetAddresses(w http.ResponseWriter, r *http.Request) {
	assetId := chi.URLParam(r, "assetId")

	addressGenerator, err := assetAddressGeneratorMap(assetId)
	if err != nil {
		// See https://developers.fireblocks.com/reference/api-responses#api-error-codes.
		writeError(w, http.StatusNotFound, "Asset doesn't exist", 1006)
		return
	}

	address, err := addressGenerator()
	if err != nil {
		log.Print(err)
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	fbAddress := fb.Address{AssetId: assetId, Address: address}
	addresses, err := json.MarshalIndent(fb.Addresses{Addresses: []fb.Address{fbAddress}}, "", "  ")
	if err != nil {
		log.Print(err)
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(utils.BinaryNewline(addresses))
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}
}

// Handler to create a new vault wallet.
// See https://developers.fireblocks.com/reference/createvaultaccountasset.
func handlePostCreateVaultAccountAsset(w http.ResponseWriter, r *http.Request) {
	// TODO: support Idempotency-Key.

	vaultAccountId := chi.URLParam(r, "vaultAccountId")
	assetId := chi.URLParam(r, "assetId")

	if vaultAccountId == "" {
		// I made this error up, it's not documented what fb would return.
		writeError(w, http.StatusBadRequest, "vaultAccountId is required", 0)
		return
	}

	addressGenerator, err := assetAddressGeneratorMap(assetId)
	if err != nil {
		// See https://developers.fireblocks.com/reference/api-responses#api-error-codes.
		writeError(w, http.StatusNotFound, "Asset doesn't exist", 1006)
		return
	}

	address, err := addressGenerator()
	if err != nil {
		log.Print(err)
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	// Seems we can get away with this as we don't need to keep track of wallet
	// IDs.
	id := strconv.Itoa(mrand.Int())

	fbVaultWallet := fb.VaultWallet{ID: id, Address: address}
	response, err := json.MarshalIndent(fbVaultWallet, "", "  ")
	if err != nil {
		log.Print(err)
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(utils.BinaryNewline(response))
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}
}

func service() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/v1/vault/accounts/{vaultAccountId}/{assetId}/addresses_paginate", handleGetAddresses)
	r.Post("/v1/vault/accounts/{vaultAccountId}/{assetId}", handlePostCreateVaultAccountAsset)
	r.Post("/v1/vault/accounts", handlePostCreateVaultAccount)
	return r
}

// Spin up the server and serve until context receives cancellation.
func RunWithCancellation(ctx context.Context, wg *sync.WaitGroup, address string) {
	server := &http.Server{Addr: address, Handler: service()}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Print("Cancelling mock server")
			if err := server.Shutdown(ctx); err != nil {
				log.Print(err)
			}
			wg.Done()
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Spin up the server and serve forever.
func Run() {
	r := service()
	address := "localhost:6200"
	log.Printf("listening on http://%s/", address)
	if err := http.ListenAndServe(address, r); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
