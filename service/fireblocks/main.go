package fb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// TODO: fix this un-DRY situation, where we've got these structs defined
// twice.

// Fireblocks vault asset, embedded in FBVaultAccount
type VaultAsset struct {
	ID            string `json:"id"`
	Total         string `json:"total"`
	Available     string `json:"available"`
	Pending       string `json:"pending"`
	Frozen        string `json:"frozen"`
	LockedAmmount string `json:"lockedAmount"`
	BlockHeight   string `json:"blockHeight"`
	BlockHash     string `json:"blockHash"`
	// RewardsInfo struct TODO
}

// Fireblocks Vault, see
// https://developers.fireblocks.com/reference/createvaultaccount.
type VaultAccount struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Assets        []VaultAsset `json:"assets"`
	HiddenOnUI    bool         `json:"hiddenOnUI"`
	CustomerRefId string       `json:"customerRefId"`
	AutoFuel      string       `json:"autoFuel"`
}

// Wallet object returned from
// https://developers.fireblocks.com/reference/createvaultaccountasset.
type VaultWallet struct {
	ID                string `json:"id"`
	Address           string `json:"address"`
	LegacyAddress     string `json:"legacyAddress,omitempty"`
	EnterpriseAddress string `json:"enterpriseAddress,omitempty"`
	Tag               string `json:"tag,omitempty"`
	EosAccountName    string `json:"eosAccountName,omitempty"`
	Status            string `json:"status,omitempty"` // TODO: use an enum.
	ActivationTxId    string `json:"activationTxId,omitempty"`
}

type Fireblocks struct {
	baseURL     string
	credentials any // Placeholder.
}

// Placeholder for Fireblocks session constructor. We'll need this to pass
// credentials through, but currently aren't concerned with that, this is just
// so we get the general shape right.
func NewFireblocksSession(baseURL string) Fireblocks {
	return Fireblocks{baseURL: baseURL}
}

func (fb *Fireblocks) CreateVaultAccount() (*VaultAccount, error) {
	createAccountURL := fb.baseURL + "/v1/vault/accounts"
	request, err := http.NewRequest(http.MethodPost, createAccountURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	var fbVaultAccount VaultAccount
	err = json.NewDecoder(response.Body).Decode(&fbVaultAccount)
	if err != nil {
		return nil, err
	}

	return &fbVaultAccount, nil
}

func (fb *Fireblocks) CreateVaultAccountAsset(accountId, assetId string) (*VaultWallet, error) {
	createAssetURL := fmt.Sprintf("%s/v1/vault/accounts/%s/%s", fb.baseURL, accountId, assetId)
	request, err := http.NewRequest(http.MethodPost, createAssetURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	var fbVaultWallet VaultWallet
	err = json.NewDecoder(response.Body).Decode(&fbVaultWallet)
	if err != nil {
		return nil, err
	}

	return &fbVaultWallet, nil
}
