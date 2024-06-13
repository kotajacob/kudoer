// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"

	"github.com/justinas/alice"
)

var rxUsername = regexp.MustCompile("^[a-z0-9_-]+$")
var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (app *application) Routes() http.Handler {
	mux := http.NewServeMux()

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /user/view/{username}", dynamic.ThenFunc(app.userView))
	mux.Handle("GET /user/register", dynamic.ThenFunc(app.userRegister))
	mux.Handle("POST /user/register", dynamic.ThenFunc(app.userRegisterPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))
	mux.Handle("GET /item/view/{id}", dynamic.ThenFunc(app.itemView))

	protected := dynamic.Append(app.requireAuthentication)

	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))
	mux.Handle("GET /item/create", protected.ThenFunc(app.itemCreate))
	mux.Handle("POST /item/create", protected.ThenFunc(app.itemCreatePost))

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