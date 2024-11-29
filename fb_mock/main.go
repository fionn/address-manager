package main

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcutil/bech32"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Fireblocks address object.
type FBAddress struct {
	AssetId           string `json:"assetId"`
	Address           string `json:"address"`
	Description       string `json:"description"`
	Tag               string `json:"tag"`
	Type              string `json:"type"`
	CustomerRefId     string `json:"customerRefId"`
	AddressFormat     string `json:"addressFormat"`
	LegacyAddress     string `json:"legacyAddress"`
	EnterpriseAddress string `json:"enterpriseAddress"`
	Bip44AddressIndex int    `json:"bip44AddressIndex"`
	UserDefined       bool   `json:"userDefined"`
}

// Fireblocks addresses object, wrapping an array of address objects.
type FBAddresses struct {
	Addresses []FBAddress `json:"addresses"`
}

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
func generateAddress() (string, error) {
	// Because we're not using a real keypair, this is just fed into some hash
	// functions. As such, we don't really care what it is, so just use some
	// random bytes.
	fakePubKey, err := randomBytes(20)
	if err != nil {
		log.Print("Failed to generate random bytes")
		return "", err
	}

	hashedPubKey := btcutil.Hash160(fakePubKey)
	witnessProgram, err := bech32.ConvertBits(hashedPubKey, 8, 5, true)
	if err != nil {
		log.Print("Failed to squash 8-bit array to 5-bit array")
		return "", err
	}

	address, err := bech32.Encode("tb", append([]byte{0}, witnessProgram...))
	if err != nil {
		log.Print("Failed to encode Bech32 address")
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

// Handler for the addresses_paginate endpoint.
func handleGetAddresses(w http.ResponseWriter, r *http.Request) {
	assetId := chi.URLParam(r, "assetId")

	// TODO: support more assets.
	if assetId != "BTC" {
		// See https://developers.fireblocks.com/reference/api-responses#api-error-codes.
		writeError(w, http.StatusNotFound, "Asset doesn't exist", 1006)
		return
	}

	address, err := generateAddress()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	fbAddress := FBAddress{AssetId: assetId, Address: address}
	addresses, err := json.MarshalIndent(FBAddresses{[]FBAddress{fbAddress}}, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), 0)
		return
	}

	// Appened a newline.
	addresses = append(addresses, []byte{10}...)
	w.Write(addresses)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/v1/vault/accounts/{vaultAccountId}/{assetId}/addresses_paginate", handleGetAddresses)

	address := "localhost:6200"
	log.Printf("listening on http://%s/", address)
	log.Fatal(http.ListenAndServe(address, r))
}
