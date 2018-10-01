package hashserver

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
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
	verbose      bool
	http         http.Server
	hashRequests chan hashRequest
	done         chan struct{}
	producers    sync.WaitGroup
	hashes       map[uint32][64]byte
	hashLock     sync.RWMutex
	numProcessed uint32
	start        time.Time
}

type hashRequest struct {
	seqNum   uint32
	password string
}

type requestStats struct {
	NumProcessed uint32 `json:"total"`
	AverageMs    uint32 `json:"average"`
	start        time.Time
}

//NewHashServer creates a new HashServer with http routes configured and ready to run
func NewHashServer(port int, verbose bool) *HashServer {
	mux := http.NewServeMux()
	srv := &HashServer{
		http: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
		verbose:      verbose,
		done:         make(chan struct{}),
		hashRequests: make(chan hashRequest, 100),
		hashes:       map[uint32][64]byte{},
	}

	mux.HandleFunc("/hash", srv.hashRequest)
	mux.HandleFunc("/hash/", srv.getHash)
	mux.HandleFunc("/stats", srv.getStats)
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		go srv.Shutdown()
		w.WriteHeader(202)
	})
	return srv
}

func (srv *HashServer) hashRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(400)
	} else {
		if pwd, ok := r.Form["password"]; !ok || len(pwd) == 0 {
			w.WriteHeader(400)
		} else {
			req := hashRequest{
				password: pwd[0],
				seqNum:   atomic.AddUint32(&srv.numProcessed, 1),
			}
			srv.producers.Add(1)
			time.AfterFunc(5*time.Second, func() {
				srv.hashRequests <- req
				srv.producers.Done()
			})
			w.WriteHeader(202)
			fmt.Fprintf(w, "%d", req.seqNum)
		}
	}
}

func (srv *HashServer) getHash(w http.ResponseWriter, r *http.Request) {
	routeParam := r.URL.Path[len("/hash/"):]
	if request, err := strconv.Atoi(routeParam); err == nil {
		srv.hashLock.RLock()
		defer srv.hashLock.RUnlock()
		if hsh, ok := srv.hashes[uint32(request)]; ok {
			w.WriteHeader(200)
			fmt.Fprintf(w, base64.StdEncoding.EncodeToString(hsh[:]))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Request: HashId \"%d\" does not exist", request)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Request: HashId \"%s\" must be a positive integer", routeParam)
	}
}

func (srv *HashServer) getStats(w http.ResponseWriter, r *http.Request) {
	stats := requestStats{NumProcessed: srv.numProcessed}
	if stats.NumProcessed > 0 {
		duration := time.Now().Sub(stats.start)
		stats.AverageMs = uint32(duration.Seconds()*1000) / stats.NumProcessed
	}

	if json, err := json.Marshal(stats); err == nil {
		i, err := w.Write(json)
		if len(json) != i || err != nil {
			log.Printf("error writing stats to ResponseWriter, only wrote %d of %d bytes, err:%v\n",
				i, len(json), err)
		}
	} else {
		w.WriteHeader(500)
		log.Printf("error marshalling stats as json response, err:%v\n", err)
	}
}

//Run the server, blocking until it's been shutdown
func (srv *HashServer) Run() error {
	srv.start = time.Now()
	go func() {
		for r := range srv.hashRequests {
			hash := sha512.Sum512([]byte(r.password))
			srv.hashLock.Lock()
			srv.hashes[r.seqNum] = hash
			srv.hashLock.Unlock()
			if srv.verbose {
				log.Printf(" -- hasing %d - %s\n", r.seqNum, base64.StdEncoding.EncodeToString(hash[:]))
			}
		}
		log.Println("ending hashing processor")
		srv.done <- struct{}{}
	}()

	log.Printf("Starting HTTP Server at %s\n", srv.http.Addr)
	if err := srv.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error running Server: %v\n", err)
		return err
	}
	<-srv.done
	return nil
}

//Shutdown the server, and wait for all pending requests to complete
func (srv *HashServer) Shutdown() {
	srv.http.Shutdown(context.Background())
	srv.producers.Wait()
	close(srv.hashRequests)
}
