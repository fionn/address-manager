# Fireblocks Mock

## Overview

This mocks the Fireblocks API. The supported endpoints are:

* [GET `/v1/vault/accounts/{vaultAccountId}/{assetId}/addresses_paginated`](https://developers.fireblocks.com/reference/getvaultaccountassetaddressespaginated),
* [POST `v1/vault/accounts`](https://developers.fireblocks.com/reference/createvaultaccount),
* [POST `v1/vault/accounts/{vaultAccountId}/{assetId}`](https://developers.fireblocks.com/reference/createvaultaccountasset).


## Usage

Run `go run ../cmd/fb_mock/main.go`, which will launch a webserver and print the address it is listening on.

An example query could be
```shell
curl -fsS http://localhost:6200/v1/vault/accounts/0/BTC/addresses_paginate | jq .
```
which would return something like
```json
{
  "addresses": [
    {
      "assetId": "BTC",
      "address": "tb1qznkaca27exg0mz6gwkcdl6t8qlagat79czzvjr",
      "description": "",
      "tag": "",
      "type": "",
      "customerRefId": "",
      "addressFormat": "",
      "legacyAddress": "",
      "enterpriseAddress": "",
      "bip44AddressIndex": 0,
      "userDefined": false
    }
  ]
}
```
where the `addresses[].address` field is a random address.
