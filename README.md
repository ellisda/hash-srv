# Go Concurrency Challenge

## Assumptions

I can assume all dev machines have docker installed or docker CLI configured for a docker environment that can be used to build / test

We'll assume we're running on linux, need build accordingly if on windows (use docker build container)

flag params for cmd

5-sec delay - can maybe still use chan if we model the 5 sec as a slow service call that must return before we can do something
 - timer.After rather than sleep. Consider garbage collection of the timer though

HTTP Server - port 8080 but configurable (golang flag, can provide env-var wrapper later if desired)
 - no or minimal muxing necesarry. Http/1, application/json only content type
 - some validation of parameters?
 - generate swagger? (not without 3rd party dependencies)
   - maybe I could include one that I generated with unsubmitted code
   - alt - write one by hand, as API docs (production code should have documented public APIs)

Graceful Failure - sigterm will begin graceful shutdown, successive sigterm ignored
 - needs stop accepting new requests, but waits for pending requests to be completed before exiting
 - should it use a non-zero exit code?
 - 

### Production Code Criteria
API Documentation - swagger, etc.

API Versioning - assume a rudimentery GET /v/status is sufficient

I need include unit tests, and ideally some integration tests (if I find that I have multiple components)

### Design Criteria
 - follow the cmd/ and pkg/ structure that I've seen in cortex

 - container build/test but build straight binary on current architecture

 - no need for dependency management (eg: dep, glide) since we're only using Std lib

 - How to integration test?