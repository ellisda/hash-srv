package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "{\"total\": 1, \"average\": 123}")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
