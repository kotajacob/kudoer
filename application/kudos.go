// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"fmt"
	"net/http"
	"strconv"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/oklog/ulid"
)

type kudoForm struct {
	Emoji int
	Body  string

	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

// kudoPostHandler creates a kudo.
func (app *application) kudoPostHandler(w http.ResponseWriter, r *http.Request) {
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

	username := app.sessionManager.GetString(r.Context(), "authenticatedUsername")
	if username == "" {
		// Shouldn't be possible due to authenticated middleware.
		app.clientError(w, http.StatusForbidden)
		return
	}

	emoji, err := strconv.Atoi(r.PostForm.Get("emoji"))
	if err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
	}
	body := r.PostForm.Get("body") // TODO: Max length 1000 characters.

	// if len(form.FieldErrors) > 0 { // TODO: form response
	// }

	_, err = app.kudos.Insert(
		r.Context(),
		itemID,
		username,
		emoji,
		body,
	)

	app.sessionManager.Put(r.Context(), "flash", "Kudos given")
	http.Redirect(w, r, fmt.Sprintf("/item/view/%v", itemID), http.StatusSeeOther)
}

// renderKudos converts the Kudo database model into the application type for
// display.
func (app *application) renderKudos(kudos []models.Kudo) []Kudo {
	var rendered []Kudo

	for _, k := range kudos {
		var r Kudo
		r.ID = k.ID.String()
		r.ItemID = k.ItemID.String()
		r.CreatorUsername = k.CreatorUsername

		switch k.Emoji {
		case 1:
			r.Emoji = "🤮"
		case 2:
			r.Emoji = "🫠"
		case 3:
			r.Emoji = "🤔"
		case 4:
			r.Emoji = "😐"
		case 5:
			r.Emoji = "🥰"
		case 6:
			r.Emoji = "🤩"
		default:
			app.errLog.Printf("kudo with invalid emoji %v: %v\n", k.Emoji, k.ID)
			continue
		}

		r.Body = k.Body

		rendered = append(rendered, r)
	}
	return rendered
}
