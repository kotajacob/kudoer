// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("GET /kudo/{id}", kudoView)
	mux.HandleFunc("GET /kudo/create", kudoCreate)
	mux.HandleFunc("POST /kudo/create", kudoCreatePost)

	log.Println("listening on :2024")

	err := http.ListenAndServe(":2024", mux)
	log.Fatalln(err)
}
