// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/oklog/ulid"
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
	uuid, err := ulid.Parse(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	kudo, err := app.kudos.Get(r.Context(), uuid)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	fmt.Fprintf(w, "%+v", kudo)
}

// kudoCreate presents a web form to add a kudo.
func (app *application) kudoCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create a kudo"))
}

// kudoCreatePost adds a kudo.
func (app *application) kudoCreatePost(w http.ResponseWriter, r *http.Request) {
	id, err := app.kudos.Insert(r.Context(), 0, "🤣", "Very funny")
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/kudo/%v", id), http.StatusSeeOther)
}
