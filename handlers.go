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

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /kudo/{id}", app.kudoView)
	mux.HandleFunc("GET /kudo/create", app.kudoCreate)
	mux.HandleFunc("POST /kudo/create", app.kudoCreatePost)

	return mux
}

// home presents a kudo.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/base.tmpl",
		"./ui/pages/home.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil)
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
func (app *application) kudoView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	msg := fmt.Sprintf("viewing kudo %d", id)
	w.Write([]byte(msg))
}

// kudoCreate presents a kudo.
func (app *application) kudoCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create a kudo"))
}

// kudoCreatePost presents a kudo.
func (app *application) kudoCreatePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("create a kudo"))
}
