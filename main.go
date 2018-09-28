package main

import (
	"flag"
	"log"

	"github.com/ellisda/hash-srv/pkg/hashserver"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	srv := hashserver.NewHashServer(*port)

	if err := srv.Run(); err != nil {
		log.Fatal("error running the hash server: %v", err)
	}
}
