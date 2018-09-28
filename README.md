# Go Concurrency Challenge

I'll not include the problem statement here publicly, but I make the following assumptions based on the requirements therein.

## Assumptions

 - I'll assume that the POSTs to `/hash/` (with trailing slash) can be considered invalid. This allows me to simplify the route muxing.
 - We'll assume we're running on on the same OS that we're building on (i.e. not cross-compiling for Windows).
 - Container-based CI/CD is an excercise for the future (i.e. no Dockerfile is requested or provided).
 - We'll use flag param for listener port, rather then environment vars (this can be easily adapted to env-vars later).
 - If we care to add rate limiting, that can be done in a load balancer or front-door authentication layer.


Graceful Failure - sigterm will begin graceful shutdown, successive sigterm ignored
 - needs stop accepting new requests, but waits for pending requests to be completed before exiting
 - should it use a non-zero exit code?
 - Q: assignment says "all remaining requests should be allowed to complete" - does this imply that we shouldn't forcefully shutdown after some timeout? (seems bad for production code to call http.Server.Shutdown without a timeout context)

 Q: What should the "stats" route return for "average" (time spent processing)? Is this suppsoed to be wall clock time spent since app start, or the summation of 5 * numRequests, or some cpu usage metric in micro-seconds? 

 Q: Can I assume that we're not trying to protect against DOS attacks and aren't interested in rate limiting? My solution to the 5-sec delay is to spawn a goroutine for every request, but this has the potential to run out of goroutines or produce an out-of-memory condition that would crash the application.

### Production Code Criteria
I've considered incoming bursts of requests, API documentation (swagger), unit tests, and code maintainability (folder structure, use of packages).

I considered API versioning, but the requirements listed routes "http://localhost:8080/hash/42" with no "/v1/" in them. I thought it overkill to look into Content-Type or X- custom headers for API versioning. I leave versioning as a future excercise.

The 5-sec delay implies a downstream asynchronous service, which would warrant integration tests. Absent further knowledge about the downstream service, I've tried to create integration test tools. That would be good to think about for test automation in a more production context.

### Design Criteria
 - follow idomatic golang folder structures for pkg/ but ignore cmd/ since there is only one program here.

 - no need for dependency management (eg: dep, glide) since we're only using Std lib

## Design Review

The 5-sec delay in the request for new hash processing could be implemented in different ways. I chose `timer.AfterFunc` which can't be cancelled or garbage collected until after it fires. Since we want the graceful shutdown to wait until the pending requests have been processed, this seemed acceptable. An alternative would be to put all incoming requests directly on a channel, but this could block the http handler if the bufffered channel filled up, which is in conflict with the requirement that the handler return immediately.


The `/stats` route returns the number of valid POSTs to `/hash`. I chose to only count successful hash requests and to exclude requests that didn't include the expected form data fields. Invalid POSTs would be another interesting metric to track.


Quick Test

`for ((i=1;i<=100;i++)); do curl -X POST "http://localhost:8080/hash" -H "accept: application/json" -H "Content-Type: application/x-www-form-urlencoded" -d "password=angryMonkey"; done`