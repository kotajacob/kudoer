// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func home(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("./ui/pages/home.tmpl")
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}
}

// kudoView presents a kudo.
func kudoView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	msg := fmt.Sprintf("viewing kudo %d", id)
	w.Write([]byte(msg))
}

// kudoCreate presents a kudo.
func kudoCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create a kudo"))
}

// kudoCreatePost presents a kudo.
func kudoCreatePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("create a kudo"))
}
