# Go Concurrency Challenge

## Assumptions

I can assume all dev machines have docker installed or docker CLI configured for a docker environment that can be used to build / test

We'll assume we're running on linux, need build accordingly if on windows (use docker build container)

flag params for cmd

5-sec delay - can maybe still use chan if we model the 5 sec as a slow service call that must return before we can do something
 - timer.After rather than sleep. Consider garbage collection of the timer though
 - Should we rate limit these? i.e. if we receive a burst of 10K hash requests in the first 5 sec, before any have completed their 5-sec wait, should we expect all 10K to succeed. How about 10M requests in first 5 sec (stored in memory)?

HTTP Server - port 8080 but configurable (golang flag, can provide env-var wrapper later if desired)
 - no or minimal muxing necesarry. Http/1, application/json only content type
 - some validation of parameters?
 - generate swagger? (not without 3rd party dependencies)
   - maybe I could include one that I generated with unsubmitted code
   - alt - write one by hand, as API docs (production code should have documented public APIs)

Graceful Failure - sigterm will begin graceful shutdown, successive sigterm ignored
 - needs stop accepting new requests, but waits for pending requests to be completed before exiting
 - should it use a non-zero exit code?
 - Q: assignment says "all remaining requests should be allowed to complete" - does this imply that we shouldn't forcefully shutdown after some timeout? (seems bad for production code to call http.Server.Shutdown without a timeout context)


 Q: Is it OK to reject POSTs to "/hash/" (i.e. with the trailing slash)? It will simplify the routing logic and allow me to specify one handler fun for GETs "/hash/" and a seperate handler func for POST "/hash".

 Q: What should the "stats" route return for "average" (time spent processing)? Is this suppsoed to be wall clock time spent since app start, or the summation of 5 * numRequests, or some cpu usage metric in micro-seconds? 

 Q: Can I assume that we're not trying to protect against DOS attacks and aren't interested in rate limiting? My solution to the 5-sec delay is to spawn a goroutine for every request, but this has the potential to run out of goroutines or produce an out-of-memory condition that would crash the application.

### Production Code Criteria
API Documentation - swagger, etc.

API Versioning - assume a rudimentery GET /v/status is sufficient

I need include unit tests, and ideally some integration tests (if I find that I have multiple components)

### Design Criteria
 - follow the cmd/ and pkg/ structure that I've seen in cortex

 - container build/test but build straight binary on current architecture

 - no need for dependency management (eg: dep, glide) since we're only using Std lib

 - How to integration test?


## Design Review

The 5-sec delay in the request for new hash processing could be implemented in different ways. I chose `timer.AfterFunc` which can't be cancelled or garbage collected until after it fires. Since we want the graceful shutdown to wait until the pending requests have been processed, this seemed acceptable. An alternative would be to put all incoming requests directly on a channel, but this could block the http handler if the bufffered channel filled up, which is in conflict with the requirement that the handler return immediately.