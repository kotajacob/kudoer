// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"git.sr.ht/~kota/kudoer/application/emoji"
	"git.sr.ht/~kota/kudoer/models"
	"github.com/oklog/ulid"
)

type itemViewPage struct {
	Page
	models.Item

	Emojis []emoji.Emoji

	// Has the user already given kudos for this item?
	Kudoed bool

	// All kudos given to this item.
	Kudos []models.Kudo
}

// itemViewHandler presents a item.
func (app *application) itemViewHandler(w http.ResponseWriter, r *http.Request) {
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

	kudos, err := app.kudos.ItemAll(r.Context(), uuid)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var kudoed bool
	if username := app.sessionManager.GetString(r.Context(), "authenticatedUsername"); username != "" {
		if _, err := app.kudos.ItemUser(r.Context(), uuid, username); errors.Is(err, models.ErrNoRecord) {
			kudoed = true
		}
	}

	title := item.Name + " - " + "Kudoer"
	app.render(w, http.StatusOK, "itemView.tmpl", itemViewPage{
		Page:   app.newPage(r, title, item.Description),
		Item:   item,
		Emojis: emoji.List(),
		Kudoed: kudoed,
		Kudos:  kudos,
	})
}

type itemCreatePage struct {
	Page
	Form itemCreateForm
}

// itemCreateHandler presents a web form to add an item.
func (app *application) itemCreateHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "itemCreate.tmpl", itemCreatePage{
		Page: app.newPage(r, "Create an item", "Create a new item on Kudoer"),
		Form: itemCreateForm{},
	})
}

type itemCreateForm struct {
	Name        string
	Description string
	Image       string

	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

// itemCreatePostHandler adds an item.
func (app *application) itemCreatePostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := itemCreateForm{
		Name:        r.PostForm.Get("name"),
		Description: r.PostForm.Get("description"),
		Image:       r.PostForm.Get("image"),
		FieldErrors: map[string]string{},
	}

	if strings.TrimSpace(form.Name) == "" {
		form.FieldErrors["name"] = "Name cannot be blank"
	} else if utf8.RuneCountInString(form.Name) > 100 {
		form.FieldErrors["name"] = "Name cannot be longer than 100 characters"
	}

	if strings.TrimSpace(form.Description) == "" {
		form.FieldErrors["description"] = "Description cannot be blank"
	} else if utf8.RuneCountInString(form.Description) > 1000 {
		form.FieldErrors["description"] = "Description cannot be longer than 1000 characters"
	}

	if len(form.FieldErrors) > 0 {
		app.render(w, http.StatusUnprocessableEntity, "itemCreate.tmpl", itemCreatePage{
			Page: app.newPage(r, "Create an item", "Create a new item on Kudoer"),
			Form: form,
		})
		return
	}

	username := app.sessionManager.GetString(r.Context(), "authenticatedUsername")
	id, err := app.items.Insert(
		r.Context(),
		username,
		form.Name,
		form.Description,
		form.Image,
	)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Item created")

	http.Redirect(w, r, fmt.Sprintf("/item/view/%v", id), http.StatusSeeOther)
}
