# Go Concurrency Challenge

Build with

    `go build`

Test with

    `go test ./...`

Take a look at the swagger.yaml in an editor like https://editor.swagger.io/

Quick System Test (while server is running)

    `for ((i=1;i<=100;i++)); do \
        curl -X POST "http://localhost:8080/hash" -d "password=angryMonkey" \
            -H "Content-Type: application/x-www-form-urlencoded"; \
    done`

## Assumptions
I've omitted the problem statement, but I make the following assumptions from it:

 - POSTs to `/hash/` (with trailing slash) may be considered invalid attempts to query `/hash/{hashId}`. This allows me to simplify the route muxing.
 - We'll use a golang flag param for listener port and default to 8080.
 - Rate limiting is outside the scope of the problem statement and might be placed in a load balancer or front-door authentication service that sits in front of this service (to be re-usable and not implemented in each service).
 - The `/stats` response defines "total" as the number of successful POSTs to `/hash`, excluding invalid requests, and "average" as the total runtime in ms divided by "total" (i.e. ms/request). This was my best interpretation of how the example "average: 123" for "total: 1"

### Production Code Criteria
I've considered incoming bursts of requests and chose not to use a buffered channel where I'd have to specify max buffer size. I've provided API documentation in the form of hand-written swagger.yaml. I've written some unit tests and follow an idiomatic "pkg/" folder structure for packages to support code maintainability.

The requirements stated not to use 3rd party packages, so there is no need for a dependency management tool like glide or dep.

I considered API versioning, but the requirements listed routes "http://localhost:8080/hash/42" with no "/v1/" in the URL. I thought it overkill to look into Content-Type or X- custom headers for API versioning. API versioning is of course an important topic for public APIs.

The 5-sec delay implies a downstream asynchronous service, which would warrant integration tests in a real production service.

## Design Review

The 5-sec delay in the request for new hash processing could be implemented in different ways. I chose `timer.AfterFunc` which can't be cancelled or garbage collected until after it fires, but we intend to wait for them all during shutdown anyway, so this is OK. An alternative would be to put all incoming requests directly on a channel, but this could block the http handler if the buffered channel filled up.

The `/stats` route returns the number of ***valid*** POSTs to `/hash`. Invalid POSTs would be another interesting metric to track.