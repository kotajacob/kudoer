// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"unicode/utf8"

	"git.sr.ht/~kota/kudoer/application/emoji"
	"git.sr.ht/~kota/kudoer/models"
	"github.com/oklog/ulid"
)

// kudoPostHandler creates a kudo.
func (app *application) kudoPostHandler(w http.ResponseWriter, r *http.Request) {
	username := app.sessionManager.GetString(r.Context(), "authenticatedUsername")
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	itemID, err := ulid.Parse(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var fieldError string
	e, err := strconv.Atoi(r.PostForm.Get("emoji"))
	if err != nil {
		fieldError = "Invalid emoji payload"
	}
	if _, err := emoji.Value(e); err != nil {
		fieldError = "Invalid emoji selected"
	}

	body := r.PostForm.Get("body")
	if utf8.RuneCountInString(body) > 1000 {
		fieldError = "Body of kudo cannot be longer than 1000 characters"
	}

	if fieldError != "" {
		app.sessionManager.Put(r.Context(), "flash", fieldError)
		http.Redirect(w, r, fmt.Sprintf("/item/view/%v", itemID), http.StatusSeeOther)
		return
	}

	k, err := app.kudos.ItemUser(
		r.Context(),
		itemID,
		username,
	)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			_, err = app.kudos.Insert(
				r.Context(),
				itemID,
				username,
				e,
				body,
			)
			app.sessionManager.Put(r.Context(), "flash", "Kudos given")
		} else {
			app.serverError(w, err)
			return
		}
	}

	err = app.kudos.Update(
		r.Context(),
		k.ID,
		itemID,
		username,
		e,
		body,
	)
	app.sessionManager.Put(r.Context(), "flash", "Kudos updated")
	http.Redirect(w, r, fmt.Sprintf("/item/view/%v", itemID), http.StatusSeeOther)
}
