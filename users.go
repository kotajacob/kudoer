// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"errors"
	"fmt"
	"net/http"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/oklog/ulid"
)

type userViewPage struct {
	CSPNonce string

	Name string
}

// userView presents a user.
func (app *application) userView(w http.ResponseWriter, r *http.Request) {
	uuid, err := ulid.Parse(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user, err := app.users.Get(r.Context(), uuid)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, http.StatusOK, "userView.tmpl", userViewPage{
		CSPNonce: nonce(r.Context()),
		Name:     user.Name,
	})
}

type userCreatePage struct {
	CSPNonce string
}

// userCreate presents a web form to add a user.
func (app *application) userCreate(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "userCreate.tmpl", userCreatePage{
		CSPNonce: nonce(r.Context()),
	})
}

// userCreatePost adds a user.
func (app *application) userCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	name := r.PostForm.Get("name")
	if name == "" {
		app.clientError(w, http.StatusBadRequest)
	}

	email := r.PostForm.Get("email")
	if email == "" {
		app.clientError(w, http.StatusBadRequest)
	}

	id, err := app.users.Insert(r.Context(), name, email)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/user/%v", id), http.StatusSeeOther)
}
