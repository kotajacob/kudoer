// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"net/http"

	"git.sr.ht/~kota/kudoer/models"
)

type searchPage struct {
	Page

	Query string
	Items []models.SearchItem
}

// searchHandler presents the item / user search page.
func (app *application) searchHandler(w http.ResponseWriter, r *http.Request) {
	data := searchPage{
		Page: app.newPage(r),
	}
	params := r.URL.Query()
	data.Query = params.Get("q")
	switch params.Get("type") {
	case "items":
		items, err := app.search.Items(r.Context(), data.Query)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data.Items = items
	}
	app.render(w, http.StatusOK, "search.tmpl", data)
}
