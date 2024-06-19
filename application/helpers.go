// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
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

// clientError sends a basic error response to the client.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// authenticated returns the session's authenticated username or a blank string
// if the user is not authenticated.
func (app *application) authenticated(r *http.Request) string {
	return app.sessionManager.GetString(r.Context(), "authenticatedUsername")
}

// strip removes any characters which are not letters, numbers, or space.
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
