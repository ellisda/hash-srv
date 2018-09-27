package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	srv := server{
		http: http.Server{
			Addr:    fmt.Sprintf(":%d", *port),
			Handler: mux,
		},
		hashRequests: make(chan hashRequest, 100),
	}

	seqNum := uint32(0)
	mux.HandleFunc("/hash", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(400)
			// fmt.Fprint(w, "unable to parse form data")
		} else {
			if pwd, ok := r.Form["password"]; !ok || len(pwd) == 0 {
				w.WriteHeader(400)
			} else {
				req := hashRequest{
					password: pwd[0], //? more than one value
					seqNum:   atomic.AddUint32(&seqNum, 1),
				}
				//no garbage collection of timer
				time.AfterFunc(5*time.Second, func() { srv.hashRequests <- req })
				w.WriteHeader(202)
				fmt.Fprintf(w, "%d", req.seqNum)
			}
		}
	})
	mux.HandleFunc("/hash/", func(w http.ResponseWriter, r *http.Request) {
		routeParam := r.URL.Path[len("/hash/"):]

		if request, err := strconv.Atoi(routeParam); err == nil {
			fmt.Fprintf(w, "Received request for Hash Num: %d", request)
		} else {
			fmt.Fprintf(w, "Invalid URL - Request \"%s\" must be a positive integer", routeParam)
		}
	})

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "{\"total\": 1, \"average\": 123}")
	})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		go srv.http.Shutdown(context.Background())
		close(srv.hashRequests)
		w.WriteHeader(202)
	})

	log.Printf("Starting HTTP Server at %s\n", srv.http.Addr)
	if err := srv.http.ListenAndServe(); err != nil {
		log.Fatalf("error running Server: %v\n", err)
	}
}

type server struct {
	http         http.Server
	hashRequests chan hashRequest
}

type hashRequest struct {
	seqNum   uint32
	password string
}
