// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"net/http"
	"strings"

	"git.sr.ht/~kota/kudoer/models"
)

type searchPage struct {
	Page
	Items []models.SearchItem

	Form searchForm
}

type searchForm struct {
	Query string

	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

// searchHandler presents the item / user search page.
func (app *application) searchHandler(w http.ResponseWriter, r *http.Request) {
	title := "Kudoer"
	params := r.URL.Query()
	form := searchForm{
		Query:       strip(params.Get("q")),
		FieldErrors: map[string]string{},
	}

	if params.Has("q") {
		if strings.TrimSpace(form.Query) == "" {
			form.FieldErrors["query"] = "Please enter a search term"
		} else {
			title = form.Query + " - " + title
		}
	}

	if len(form.FieldErrors) > 0 {
		app.render(w, http.StatusUnprocessableEntity, "search.tmpl", searchPage{
			Page: app.newPage(r, title, "Search Kudoer for items to review!"),
			Form: form,
		})
		return
	}

	var items []models.SearchItem
	switch params.Get("type") {
	case "items":
		i, err := app.search.Items(r.Context(), form.Query)
		if err != nil {
			app.serverError(w, err)
			return
		}
		items = i
	}
	app.render(w, http.StatusOK, "search.tmpl",
		searchPage{
			Page:  app.newPage(r, title, "Search Kudoer for items to review!"),
			Items: items,
			Form:  form,
		})
}
