# Address Manager

## Build

To build the mock Fireblocks server and the address manager service, use `make bin/fb_mock` and `make bin/service` (respectively).

## Test

### Automated Tests

Unit tests using the mock exist in `service/` and can be run with `make test`.

### Manual Tests

To run _ad hoc_ tests, you must run the mock and service. First build (see above) and run (by executing the binaries) both servers, then send HTTP requests to them.

See [`service/`](service/) and [`fb_mock/`](fb_mock/) for documentation on what endpoints they serve.

## Future Work

This is an MVP and not "production ready". Follow-up work:

* [ ] if using PostgreSQL, `u.BeforeCreate` can be dropped as we get native UUID support,
* [ ] move `fireblocks` out of `service` as it's common between it and `fb_mock`, _or_ move `fb_mock` into service as a dedicated part of its test suite,
* [ ] more and better unit tests:
  * test more granularly,
  * test `service`'s API endpoints themselves, not just the functions underneath,
  * test for negative conditions, not just the happy path,
  * etc.,
* [ ] use exponential backoff,
* [ ] explore the possibility of using caching,
* [ ] make the JSON structure returned by `service` nicer (at the very least use snake case),
* [ ] improve error handling (e.g. sentinel values), this was a little rushed,
* [ ] improve logging (the standard library logger doesn't support levels or structured logs).
