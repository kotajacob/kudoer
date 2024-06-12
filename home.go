package main

import "net/http"

type homePage struct {
	CSPNonce string
	Flash    string
}

// home presents the home page.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	flash := app.sessionManager.PopString(r.Context(), "flash")
	app.render(w, http.StatusOK, "home.tmpl", homePage{
		CSPNonce: nonce(r.Context()),
		Flash:    flash,
	})
}
