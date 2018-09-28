package hashserver

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

//HashServer is an HTTP server that processes password hasing requests
type HashServer struct {
	http         http.Server
	hashRequests chan hashRequest
	hashes       map[uint32][64]byte
	hashLock     sync.RWMutex
	seqNum       uint32
}

type hashRequest struct {
	seqNum   uint32
	password string
}

//NewHashServer creates a new HashServer with http routes configured and ready to run
func NewHashServer(port int) *HashServer {
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

	mux.HandleFunc("/hash", srv.hashRequest)
	mux.HandleFunc("/hash/", srv.getHash)
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "{\"total\": 1, \"average\": 123}")
	})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		go srv.Shutdown()
		w.WriteHeader(202)
	})
	return srv
}

func (srv *HashServer) hashRequest(w http.ResponseWriter, r *http.Request) {
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
			time.AfterFunc(5*time.Second, func() { srv.hashRequests <- req })
			w.WriteHeader(202)
			fmt.Fprintf(w, "%d", req.seqNum)
		}
	}
}

func (srv *HashServer) getHash(w http.ResponseWriter, r *http.Request) {
	routeParam := r.URL.Path[len("/hash/"):]

	if request, err := strconv.Atoi(routeParam); err == nil {
		// fmt.Fprintf(w, "Received request for Hash Num: %d", request)
		srv.hashLock.RLock()
		defer srv.hashLock.RUnlock()
		if hsh, ok := srv.hashes[uint32(request)]; ok {
			w.WriteHeader(200)
			fmt.Fprintf(w, base64.StdEncoding.EncodeToString(hsh[:]))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid URL - Request \"%d\" does not exist", request)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid URL - Request \"%s\" must be a positive integer", routeParam)
	}
}

//Run the server, blocking until it's been shutdown
func (srv *HashServer) Run() error {
	go func() {
		for r := range srv.hashRequests {
			hash := sha512.Sum512([]byte(r.password))
			srv.hashLock.Lock()
			srv.hashes[r.seqNum] = hash
			srv.hashLock.Unlock()
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
