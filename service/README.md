# Address Manager Service

## Overview

Service that allocates wallet addresses to users, with the following properties:
* maintain a pool of pre-allocated addresses, to quickly allocate without blocking on Fireblocks API calls,
* manage customer records statefully such that we can survive a restart,
* expose a REST API for:
  * creating users (and allocating addresses to them),
  * fetching users.

## Testing

### Unit Tests

```shell
go test -v
```

### Ad Hoc Tests

Run the Fireblocks mock server in `../fb_mock`, and point this service and the mock's URL.

Run this service (with e.g. `go run ../cmd/service/main.go`).

The supported endpoints are:
* POST `/user` to create a user, returns user data as a JSON blob,
* GET `/user/{userId}` to get a user with a given ID, returns the same user data.

<details>
<summary>Example</summary>

To create a user,
```shell
curl -X POST http://localhost:6201/user
```
which would return
```json
{
  "ID": "3f2b3ec2-44e2-4075-b91e-e17203e9938a",
  "CreatedAt": "2025-01-31T01:58:44.543023+08:00",
  "UpdatedAt": "2025-01-31T01:58:44.543023+08:00",
  "DeletedAt": null,
  "Wallet": {
    "ID": 1,
    "CreatedAt": "2025-01-31T01:58:44.543306+08:00",
    "UpdatedAt": "2025-01-31T01:58:44.543306+08:00",
    "DeletedAt": null,
    "AddressBTC": "tb1qskvstafcxuztc9jl53c4jcujqkfux6pprlgsr3",
    "UserID": "3f2b3ec2-44e2-4075-b91e-e17203e9938a"
  }
}
```
or similar.

To fetch a user given their user ID,
```shell
curl http://localhost:6201/user/3f2b3ec2-44e2-4075-b91e-e17203e9938a
```
which would return
```json
{
  "ID": "3f2b3ec2-44e2-4075-b91e-e17203e9938a",
  "CreatedAt": "2025-01-31T01:58:44.543023+08:00",
  "UpdatedAt": "2025-01-31T01:58:44.543023+08:00",
  "DeletedAt": null,
  "Wallet": {
    "ID": 1,
    "CreatedAt": "2025-01-31T01:58:44.543306+08:00",
    "UpdatedAt": "2025-01-31T01:58:44.543306+08:00",
    "DeletedAt": null,
    "AddressBTC": "tb1qskvstafcxuztc9jl53c4jcujqkfux6pprlgsr3",
    "UserID": "3f2b3ec2-44e2-4075-b91e-e17203e9938a"
  }
}
```
or similar (i.e. the same response as user creation).
</details>
