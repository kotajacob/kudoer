package main

import (
	"errors"
	"fmt"
	"net/http"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/oklog/ulid"
)

type kudoViewPage struct {
	CSPNonce string

	ID string
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
			app.serverError(w, err)
		}
		return
	}

	app.render(w, http.StatusOK, "kudoView.tmpl", kudoViewPage{
		CSPNonce: nonce(r.Context()),
		ID:       kudo.ID.String(),
	})
}

// kudoCreate presents a web form to add a kudo.
func (app *application) kudoCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create a kudo"))
}

// kudoCreatePost adds a kudo.
func (app *application) kudoCreatePost(w http.ResponseWriter, r *http.Request) {
	id, err := app.kudos.Insert(r.Context(), 0, "🤣", "Very funny")
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/kudo/%v", id), http.StatusSeeOther)
}
