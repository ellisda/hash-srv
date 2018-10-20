# Go Concurrency Challenge
[![Build Status](https://travis-ci.org/ellisda/hash-srv.svg?branch=master)](https://travis-ci.org/ellisda/hash-srv)

Build with

    go build

Test with

    go test ./...

Take a look at the swagger.yaml in an editor like https://editor.swagger.io/

Run (with verbose logging)

    go build -o hash-srv && ./hash-srv -v

Quick System Test (while server is running)

    for ((i=1;i<=100;i++)); do \
        curl -X POST "http://localhost:8080/hash" -d "password=angryMonkey" \
            -H "Content-Type: application/x-www-form-urlencoded";
    done > /dev/null 2>&1

Docker Build / Deploy (from [Building Minimal Docker Containers for Go Applications](https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/))

    docker build -t hash-srv -f Dockerfile .

    docker run -p 8080:8080 hash-srv


## Assumptions
I've omitted the problem statement, but I make the following assumptions from it:

 - POSTs to `/hash/` (with trailing slash) may be considered invalid (i.e. invalid attempts to query `/hash/{hashId}`). This allows me to simplify the route muxing.
 - We'll listen on port 8080 by default - adjustable with cmd line flag.
 - Rate limiting is outside the scope of the problem statement. That responsibility might better be implemented in an API gateway or authentication service that sits in front of this service (to be re-usable and not implemented in each service).
 - My interpretation of the `/stats` response is that "total" is the number of successful POSTs to `/hash`, excluding invalid requests, and that "average" is the average time spent in the POST `/hash` handler (excluding the 5 sec delay and the actual hash processing).
 
### Production Code Considerations
To round this out, I've provided a swagger.yaml for API documentation. I've written some unit tests and followed an idiomatic "pkg/" folder structure for code maintainability.

The requirements stated not to use 3rd party packages, so there is no need for a dependency management tool like glide or dep.

I considered API versioning, but the requirements listed routes "http://localhost:8080/hash/42" with no "/v1/" in the URL. I thought it overkill to look into Content-Type or X- custom headers for API versioning. API versioning is of course an important topic for public APIs.

The 5-sec delay implies a downstream asynchronous service, which would warrant integration tests in a real production service.

The `/stats` route returns the number of ***valid*** POSTs to `/hash`. Invalid POSTs would be another interesting metric to track.

## Design Review
To gracefully handle bursts of requests, I chose timers rather than directly inserting requests into a buffered channel where I'd have to specify max buffer size.

The 5-sec delay in the request for new hash processing could be implemented in different ways. I chose `timer.AfterFunc` which can't be cancelled or garbage collected until after it fires, but we intend to wait for them all during shutdown anyway, so this is OK. An alternative would be to put all incoming requests directly on a channel, but this could block the http handler if the buffered channel filled up.
