package fb_mock

import (
	"crypto/rand"
	"encoding/json"
	"log"
	mrand "math/rand"
	"net/http"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcutil/bech32"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	fb "github.com/fionn/address-manager/service/fireblocks"
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

// Append a byte-decoded newline to a bytearray.
// This helper function exists only to add a descriptive name to this common
// operation.
func binaryNewline(s []byte) []byte {
	return append(s, []byte{10}...)
}

// Generate a random Bech32 address.
func generateBTCAddress() (string, error) {
	// Because we're not using a real keypair, this is just fed into some hash
	// functions. As such, we don't really care what it is, so just use some
	// random bytes.
	fakePubKey, err := randomBytes(20)
	if err != nil {
		log.Printf("Failed to generate random bytes: %s", err)
		return "", err
	}

	hashedPubKey := btcutil.Hash160(fakePubKey)
	witnessProgram, err := bech32.ConvertBits(hashedPubKey, 8, 5, true)
	if err != nil {
		log.Printf("Failed to squash 8-bit array to 5-bit array: %s", err)
		return "", err
	}

	address, err := bech32.Encode("tb", append([]byte{0}, witnessProgram...))
	if err != nil {
		log.Printf("Failed to encode Bech32 address: %s", err)
		return "", err
	}

	return address, nil
}

// Helper to write error messages as HTTP responses.
func writeError(w http.ResponseWriter, httpErrorCode int, message string, apiErrorCode int) {
	fbError, _ := json.MarshalIndent(FBError{apiErrorCode, message}, "", "  ")
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
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	_, err = w.Write(binaryNewline(response))
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}
}

// Handler for the addresses_paginate endpoint.
// See https://developers.fireblocks.com/reference/getvaultaccountassetaddressespaginated.
func handleGetAddresses(w http.ResponseWriter, r *http.Request) {
	assetId := chi.URLParam(r, "assetId")

	// TODO: support more assets.
	if assetId != "BTC" {
		// See https://developers.fireblocks.com/reference/api-responses#api-error-codes.
		writeError(w, http.StatusNotFound, "Asset doesn't exist", 1006)
		return
	}

	address, err := generateBTCAddress()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	fbAddress := fb.Address{AssetId: assetId, Address: address}
	addresses, err := json.MarshalIndent(fb.Addresses{Addresses: []fb.Address{fbAddress}}, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	_, err = w.Write(binaryNewline(addresses))
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

	// TODO: support more assets.
	if assetId != "BTC" {
		// See https://developers.fireblocks.com/reference/api-responses#api-error-codes.
		writeError(w, http.StatusNotFound, "Asset doesn't exist", 1006)
		return
	}

	address, err := generateBTCAddress()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	// Seems we can get away with this as we don't need to keep track of wallet
	// IDs.
	id := strconv.Itoa(mrand.Int())

	fbVaultWallet := fb.VaultWallet{ID: id, Address: address}
	response, err := json.MarshalIndent(fbVaultWallet, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	_, err = w.Write(binaryNewline(response))
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

// Spin up webserver and serve at address forever.
func RunWithAddress(address string) {
	// TODO: consider having this be context-aware and support cancellation.
	server := &http.Server{Addr: address, Handler: service()}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// Spin up webserver and serve forever.
func Run() {
	r := service()
	address := "localhost:6200"
	log.Printf("listening on http://%s/", address)
	if err := http.ListenAndServe(address, r); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
