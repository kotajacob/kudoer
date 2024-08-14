// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"

	"git.sr.ht/~kota/kudoer/ui"
	"github.com/justinas/alice"
)

func (app *application) Routes() http.Handler {
	mux := http.NewServeMux()

	media := http.FileServer(http.Dir(app.mediaStore.Dir()))
	mux.Handle("GET /media/", immutable(http.StripPrefix("/media", media)))
	mux.Handle("GET /static/", app.FromHash(immutable(http.FileServerFS(ui.EFS))))

	subFS, err := fs.Sub(ui.EFS, "static")
	if err != nil {
		app.errLog.Fatal(err) // Should be impossible with embedded FS.
	}
	mux.Handle("GET /robots.txt", http.FileServerFS(subFS))

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.homeHandler))
	mux.Handle("GET /all", dynamic.ThenFunc(app.allHandler))
	mux.Handle("GET /search", dynamic.ThenFunc(app.searchHandler))
	mux.Handle("GET /user/view/{username}", dynamic.ThenFunc(app.userViewHandler))
	mux.Handle("GET /user/followers/{username}", dynamic.ThenFunc(app.userFollowersHandler))
	mux.Handle("GET /user/following/{username}", dynamic.ThenFunc(app.userFollowingHandler))
	mux.Handle("GET /user/register", dynamic.ThenFunc(app.userRegisterHandler))
	mux.Handle("POST /user/register", dynamic.ThenFunc(app.userRegisterPostHandler))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLoginHandler))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPostHandler))
	mux.Handle("GET /user/forgot", dynamic.ThenFunc(app.userForgotHandler))
	mux.Handle("POST /user/forgot", dynamic.ThenFunc(app.userForgotPostHandler))
	mux.Handle("GET /user/reset", dynamic.ThenFunc(app.userResetHandler))
	mux.Handle("POST /user/reset", dynamic.ThenFunc(app.userResetPostHandler))
	mux.Handle("GET /item/view/{id}", dynamic.ThenFunc(app.itemViewHandler))

	protected := dynamic.Append(app.requireAuthentication)

	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPostHandler))
	mux.Handle("GET /user/settings", protected.ThenFunc(app.userSettingsHandler))
	mux.Handle("POST /user/settings", protected.ThenFunc(app.userSettingsPostHandler))
	mux.Handle("POST /user/follow", protected.ThenFunc(app.userFollowPostHandler))
	mux.Handle("POST /user/unfollow", protected.ThenFunc(app.userUnfollowPostHandler))
	mux.Handle("GET /item/create", protected.ThenFunc(app.itemCreateHandler))
	mux.Handle("POST /item/create", protected.ThenFunc(app.itemCreatePostHandler))
	mux.Handle("POST /kudo/{id}", protected.ThenFunc(app.kudoPostHandler))

	standard := alice.New(
		app.recoverPanic,
		app.rateLimiter.RateLimit,
		app.logRequest,
		app.secureHeaders,
	)
	return standard.Then(mux)
}

func (app *application) render(
	w http.ResponseWriter,
	status int,
	page string,
	data any,
) {
	ts, ok := app.templates[page]
	if !ok {
		app.serverError(w, fmt.Errorf(
			"the template %s is missing",
			page,
		))
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}
