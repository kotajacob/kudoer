// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /user/{id}", app.userView)
	mux.HandleFunc("GET /user/create", app.userCreate)
	mux.HandleFunc("POST /user/create", app.userCreatePost)
	mux.HandleFunc("GET /item/{id}", app.itemView)
	mux.HandleFunc("GET /item/create", app.itemCreate)
	mux.HandleFunc("POST /item/create", app.itemCreatePost)
	mux.HandleFunc("GET /kudo/{id}", app.kudoView)
	mux.HandleFunc("GET /kudo/create", app.kudoCreate)
	mux.HandleFunc("POST /kudo/create", app.kudoCreatePost)

	standard := alice.New(app.recoverPanic, app.logRequest, app.secureHeaders)

	return standard.Then(mux)
}

func (app *application) render(
	w http.ResponseWriter,
	status int,
	page string,
	data interface{},
) {
	ts, ok := app.templates[page]
	if !ok {
		app.serverError(w, fmt.Errorf(
			"the template %s is missing",
			page,
		))
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

type homePage struct {
	CSPNonce string
}

// home presents the home page.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "home.tmpl", homePage{
		CSPNonce: nonce(r.Context()),
	})
}
