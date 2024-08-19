// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
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

// page checks for the page URL parameter and returns a valid page number.
func page(params url.Values) int {
	if ok := params.Has("page"); ok {
		page, err := strconv.Atoi(params.Get("page"))
		if err == nil && page > 1 {
			return page
		}
	}
	return 1
}

// destroySessions will remove every session for a given username.
// This logs the user out on all of their computers.
func (app *application) destroySessions(username string) error {
	ctx := context.WithValue(context.Background(), ContextKeyUsername, username)
	fn := func(ctx context.Context) error {
		want := ctx.Value(ContextKeyUsername)
		got := app.sessionManager.Get(ctx, "authenticatedUsername")
		if want == got {
			return app.sessionManager.Destroy(ctx)
		}
		return nil
	}
	return app.sessionManager.Iterate(ctx, fn)
}

// login will authenticate the current session as the provided user.
func (app *application) login(
	r *http.Request,
	username string,
	rememberMe bool,
) error {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}
	app.sessionManager.Put(r.Context(), "authenticatedUsername", username)
	if rememberMe {
		app.sessionManager.SetDeadline(
			r.Context(),
			time.Now().Add(time.Hour*24*365*10),
		)
	}
	app.sessionManager.RememberMe(r.Context(), rememberMe)
	return nil
}

// logout will dis-authenticate the current session.
func (app *application) logout(r *http.Request) error {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUsername")
	return nil
}

// flash will show a given message on the next page load for the current
// session.
func (app *application) flash(r *http.Request, msg string) {
	app.sessionManager.Put(r.Context(), "flash", msg)
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
