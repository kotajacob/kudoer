// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"errors"
	"fmt"
	"net/http"

	"git.sr.ht/~kota/kudoer/application/validator"
	"git.sr.ht/~kota/kudoer/db/models"
	"github.com/oklog/ulid"
)

// kudoPostHandler creates a kudo.
func (app *application) kudoPostHandler(w http.ResponseWriter, r *http.Request) {
	username := app.authenticated(r)
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

	v := validator.New()
	e, f, body := v.Kudo(
		r.PostForm.Get("emoji"),
		r.PostForm.Get("frame"),
		r.PostForm.Get("body"),
	)

	_, fieldErrors, valid := v.Valid()
	if !valid {
		app.flash(r, fieldErrors["kudo"])
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
			if _, err := app.kudos.Insert(
				r.Context(),
				itemID,
				username,
				f,
				e,
				body,
			); err != nil {
				app.serverError(w, err)
				return
			}
			app.flash(r, "Kudos given")
		} else {
			app.serverError(w, err)
			return
		}
	}

	if err := app.kudos.Update(
		r.Context(),
		k.ID,
		itemID,
		username,
		f,
		e,
		body,
	); err != nil {
		app.serverError(w, err)
		return
	}
	app.flash(r, "Kudos updated")
	http.Redirect(w, r, fmt.Sprintf("/item/view/%v", itemID), http.StatusSeeOther)
}
