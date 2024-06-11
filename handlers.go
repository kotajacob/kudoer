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

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /user/{id}", dynamic.ThenFunc(app.userView))
	mux.Handle("GET /user/create", dynamic.ThenFunc(app.userCreate))
	mux.Handle("POST /user/create", dynamic.ThenFunc(app.userCreatePost))
	mux.Handle("GET /item/{id}", dynamic.ThenFunc(app.itemView))
	mux.Handle("GET /item/create", dynamic.ThenFunc(app.itemCreate))
	mux.Handle("POST /item/create", dynamic.ThenFunc(app.itemCreatePost))

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
