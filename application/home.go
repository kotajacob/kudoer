// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import "net/http"

type homePage struct {
	Page
}

// home presents the home page.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "home.tmpl", homePage{
		Page: app.newPage(r),
	})
}
