// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("home"))
}

// kudoView presents an kudo.
func kudoView(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("view an kudo"))
}

// kudoCreate presents an kudo.
func kudoCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create an kudo"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", home)
	mux.HandleFunc("/kudo/view", kudoView)
	mux.HandleFunc("/kudo/create", kudoCreate)

	log.Println("listening on :2024")

	err := http.ListenAndServe(":2024", mux)
	log.Fatalln(err)
}
