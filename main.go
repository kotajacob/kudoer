// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":2024", "HTTP Network Address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("GET /kudo/{id}", kudoView)
	mux.HandleFunc("GET /kudo/create", kudoCreate)
	mux.HandleFunc("POST /kudo/create", kudoCreatePost)

	log.Println("listening on", *addr)

	err := http.ListenAndServe(*addr, mux)
	log.Fatalln(err)
}
