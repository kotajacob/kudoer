// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// serverError writes a log entry and then sends a generic Internal Server Error
// response to the client.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errLog.Output(2, trace)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "authenticatedUsername")
}
