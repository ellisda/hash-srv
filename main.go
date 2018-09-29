package main

import (
	"flag"
	"log"

	"github.com/ellisda/hash-srv/pkg/hashserver"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	verbose := flag.Bool("v", false, "verbose logging")
	flag.Parse()

	srv := hashserver.NewHashServer(*port, *verbose)

	if err := srv.Run(); err != nil {
		log.Fatalf("error running the hash server: %v", err)
	}
}
