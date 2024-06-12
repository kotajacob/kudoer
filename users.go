// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/oklog/ulid"
)

type userViewPage struct {
	CSPNonce string

	Username string
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
		Username: user.Username,
	})
}

type userCreatePage struct {
	CSPNonce string
	Form     userCreateForm
}

// userCreate presents a web form to add a user.
func (app *application) userCreate(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "userCreate.tmpl", userCreatePage{
		CSPNonce: nonce(r.Context()),
		Form:     userCreateForm{},
	})
}

type userCreateForm struct {
	Username    string
	Email       string
	FieldErrors map[string]string
}

// userCreatePost adds a user.
func (app *application) userCreatePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userCreateForm{
		Username:    r.PostForm.Get("username"),
		Email:       r.PostForm.Get("email"),
		FieldErrors: map[string]string{},
	}

	if strings.TrimSpace(form.Username) == "" {
		form.FieldErrors["username"] = "Username cannot be blank"
	} else if utf8.RuneCountInString(form.Username) > 30 {
		form.FieldErrors["username"] = "Username cannot be longer than 30 characters"
	}

	if len(form.Email) > 254 || !rxEmail.MatchString(form.Email) {
		// https://stackoverflow.com/questions/386294/what-is-the-maximum-length-of-a-valid-email-address
		form.FieldErrors["email"] = "Email appears to be invalid"
	}

	if len(form.FieldErrors) > 0 {
		app.render(w, http.StatusUnprocessableEntity, "userCreate.tmpl", userCreatePage{
			CSPNonce: nonce(r.Context()),
			Form:     form,
		})
		return
	}

	id, err := app.users.Insert(
		r.Context(),
		form.Username,
		form.Email,
	)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/user/%v", id), http.StatusSeeOther)
}
