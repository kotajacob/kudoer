// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"net/http"
	"strings"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/blevesearch/bleve"
)

type searchPage struct {
	Page
	Items []models.Item

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

	var items []models.Item
	switch params.Get("type") {
	case "items":
		var err error
		items, err = app.searchItems(form.Query, r)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	app.render(w, http.StatusOK, "search.tmpl",
		searchPage{
			Page:  app.newPage(r, title, "Search Kudoer for items to review!"),
			Items: items,
			Form:  form,
		})
}

func (app *application) searchItems(q string, r *http.Request) ([]models.Item, error) {
	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := app.itemSearch.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	// We take the list of IDs and their rankings and look up their names and
	// descriptions in the database for rendering the search result.
	var ids []models.SortedID
	for i, hit := range searchResult.Hits {
		ids = append(ids, models.SortedID{Index: i, ID: hit.ID})
	}

	return app.items.GetList(r.Context(), ids)
}
