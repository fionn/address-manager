# Address Manager Service

> [!NOTE]
> This doesn't exist yet, just a skeleton.

## Overview

Service that allocates wallet addresses to users, with the following properties:
* maintain a pool of pre-allocated addresses, to quickly allocate without blocking on Fireblocks API calls,
* manage customer records statefully such that we can survive a restart,
* expose a REST API for:
  * creating users
  * allocating addresses to users.

## Testing

### Unit Tests

TODO.

### Integration Tests

Run the Fireblocks mock server in `../fb_mock`, and point this service and the mock's URL.

TODO: make this a one-click setup and execute predefined integration tests.
