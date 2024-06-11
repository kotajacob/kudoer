// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/oklog/ulid"
)

type itemViewPage struct {
	CSPNonce string
	Flash    string

	Name        string
	Description string
	Image       string
}

// itemView presents a item.
func (app *application) itemView(w http.ResponseWriter, r *http.Request) {
	uuid, err := ulid.Parse(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	item, err := app.items.Get(r.Context(), uuid)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, err)
		}
		return
	}

	flash := app.sessionManager.PopString(r.Context(), "flash")

	app.render(w, http.StatusOK, "itemView.tmpl", itemViewPage{
		CSPNonce:    nonce(r.Context()),
		Flash:       flash,
		Name:        item.Name,
		Description: item.Description,
		Image:       item.Image,
	})
}

type itemCreatePage struct {
	CSPNonce string
}

// itemCreate presents a web form to add an item.
func (app *application) itemCreate(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "itemCreate.tmpl", itemCreatePage{
		CSPNonce: nonce(r.Context()),
	})
}

// itemCreatePost adds an item.
func (app *application) itemCreatePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ms := ulid.Timestamp(time.Now())
	creator_id, err := ulid.New(ms, rand.Reader)
	if err != nil {
		app.serverError(w, err)
		return
	}

	name := r.PostForm.Get("name")
	if strings.TrimSpace(name) == "" {
		app.clientError(w, http.StatusBadRequest)
	}

	description := r.PostForm.Get("description")
	if strings.TrimSpace(description) == "" {
		app.clientError(w, http.StatusBadRequest)
	}

	image := r.PostForm.Get("image")
	if strings.TrimSpace(image) == "" {
		app.clientError(w, http.StatusBadRequest)
	}

	id, err := app.items.Insert(r.Context(), creator_id, name, description, image)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "item created")

	http.Redirect(w, r, fmt.Sprintf("/item/%v", id), http.StatusSeeOther)
}
