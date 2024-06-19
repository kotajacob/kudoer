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
	Items []SearchItem

	Form searchForm
}

type SearchItem struct {
	ID          string
	Name        string
	Description string
}

type searchForm struct {
	Query string

	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

// searchHandler presents the item / user search page.
func (app *application) searchHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	form := searchForm{
		Query:       strip(params.Get("q")),
		FieldErrors: map[string]string{},
	}

	if strings.TrimSpace(form.Query) == "" && params.Has("q") {
		form.FieldErrors["query"] = "Please enter a search term"
	}

	if len(form.FieldErrors) > 0 {
		app.render(w, http.StatusUnprocessableEntity, "search.tmpl", searchPage{
			Page: app.newPage(r),
			Form: form,
		})
		return
	}

	var items []SearchItem
	switch params.Get("type") {
	case "items":
		i, err := app.search.Items(r.Context(), form.Query)
		if err != nil {
			app.serverError(w, err)
			return
		}
		items = app.renderSearchItems(i)
	}
	app.render(w, http.StatusOK, "search.tmpl",
		searchPage{
			Page:  app.newPage(r),
			Items: items,
			Form:  form,
		})
}

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

// renderSearchItems converts the SearchItem database model into the application
// type for display.
func (app *application) renderSearchItems(items []models.SearchItem) []SearchItem {
	var rendered []SearchItem

	for _, i := range items {
		var r SearchItem
		r.ID = i.ID.String()
		r.Name = i.Name
		r.Description = i.Description
		rendered = append(rendered, r)
	}
	return rendered
}
