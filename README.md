# Address Manager

## Build

To build the mock Fireblocks server and the address manager service, use `make bin/fb_mock` and `make bin/service` (respectively).

## Test

### Automated Tests

Unit tests using the mock exist in `service/` and can be run with `make test`.

### Manual Tests

To run _ad hoc_ tests, you must run the mock and service. First build (see above) and run (by executing the binaries) both servers, then send HTTP requests to them.

See [`service/`](service/) and [`fb_mock/`](fb_mock/) for documentation on what endpoints they serve.
