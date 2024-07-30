// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"net/http"

	"git.sr.ht/~kota/kudoer/db/models"
)

type homePage struct {
	Page

	// All kudos given to this item.
	Kudos []models.Kudo
}

// homeHandler presents the homeHandler page.
func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
	var kudos []models.Kudo
	username := app.authenticated(r)
	if username != "" {
		var err error
		kudos, err = app.kudos.Following(r.Context(), username)
		if err != nil {
			app.serverError(w, err)
			return
		}
	} else {
		var err error
		kudos, err = app.kudos.All(r.Context())
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	app.render(w, http.StatusOK, "home.tmpl", homePage{
		Page:  app.newPage(r, "Kudoer", "Give kudos to your favorite things!"),
		Kudos: kudos,
	})
}

// allHandler presents a page displaying kudos from all users.
// This is what the homepage displays when not logged in.
func (app *application) allHandler(w http.ResponseWriter, r *http.Request) {
	kudos, err := app.kudos.All(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, http.StatusOK, "home.tmpl", homePage{
		Page:  app.newPage(r, "Kudoer", "Give kudos to your favorite things!"),
		Kudos: kudos,
	})
}
