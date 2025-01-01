module github.com/fionn/address-manager/fb_mock

go 1.23.3

require (
	github.com/btcsuite/btcd/btcutil v1.1.6
	github.com/btcsuite/btcutil v1.0.2
	github.com/fionn/address-manager/service v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.2.0
)

require (
	github.com/btcsuite/btcd v0.24.2 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

replace github.com/fionn/address-manager/service => ../service
