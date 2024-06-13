// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import "net/http"

type homePage struct {
	Page
}

// homeHandler presents the homeHandler page.
func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "home.tmpl", homePage{
		Page: app.newPage(r),
	})
}
