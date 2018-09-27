package hashserver

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

//HashServer is an HTTP server that processes password hasing requests
type HashServer struct {
	http         http.Server
	hashRequests chan hashRequest
	hashes       map[uint32][64]byte
	seqNum       uint32
}

type hashRequest struct {
	seqNum   uint32
	password string
}

//New creates a new HashServer, with http routes configured and ready to run
func New(port int) *HashServer {
	mux := http.NewServeMux()
	srv := &HashServer{
		http: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
		hashRequests: make(chan hashRequest, 100),
		hashes:       map[uint32][64]byte{},
		seqNum:       uint32(0),
	}

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
					seqNum:   atomic.AddUint32(&srv.seqNum, 1),
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
		go srv.Shutdown()
		w.WriteHeader(202)
	})
	return srv
}

//Run the server, blocking until it's been shutdown
func (srv *HashServer) Run() error {
	go func() {
		for r := range srv.hashRequests {
			hash := sha512.Sum512([]byte(r.password))
			srv.hashes[r.seqNum] = hash
			log.Printf(" -- hasing %d - %s\n", r.seqNum, base64.StdEncoding.EncodeToString(hash[:]))
		}
		log.Println("ending hashing processor")
	}()

	log.Printf("Starting HTTP Server at %s\n", srv.http.Addr)
	if err := srv.http.ListenAndServe(); err != nil {
		log.Fatalf("error running Server: %v\n", err)
		return err
	}
	return nil
}

//Shutdown the server, and wait for all pending requests to complete
func (srv *HashServer) Shutdown() {
	srv.http.Shutdown(context.Background())
	close(srv.hashRequests)
	//TODO - wait for pending requests that are already on chan
}
