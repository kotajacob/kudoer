// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("home"))
}

// kudoView presents an kudo.
func kudoView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	msg := fmt.Sprintf("viewing kudo %d", id)
	w.Write([]byte(msg))
}

// kudoCreate presents an kudo.
func kudoCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create an kudo"))
}

// kudoCreatePost presents an kudo.
func kudoCreatePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("create an kudo"))
}

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
