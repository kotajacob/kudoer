// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

type application struct {
	infoLog        *log.Logger
	errLog         *log.Logger
	templates      map[string]*template.Template
	sessionManager *scs.SessionManager

	users  *models.UserModel
	items  *models.ItemModel
	kudos  *models.KudoModel
	search *models.SearchModel
}

func New(
	infoLog *log.Logger,
	errLog *log.Logger,
	templates map[string]*template.Template,
	sessionManager *scs.SessionManager,
	users *models.UserModel,
	items *models.ItemModel,
	kudos *models.KudoModel,
	search *models.SearchModel,
) *application {
	return &application{
		infoLog:        infoLog,
		errLog:         errLog,
		templates:      templates,
		sessionManager: sessionManager,
		users:          users,
		items:          items,
		kudos:          kudos,
		search:         search,
	}
}

// Page represents basic information needed on every page.
type Page struct {
	CSPNonce        string
	CSRFToken       string
	Flash           string
	Authenticated   string
	Title           string
	PageDescription string
}

func (app *application) newPage(r *http.Request, title, description string) Page {
	cspNonce := nonce(r.Context())
	csrfToken := nosurf.Token(r)
	flash := app.sessionManager.PopString(r.Context(), "flash")
	authenticated := app.authenticated(r)
	return Page{
		CSPNonce:        cspNonce,
		CSRFToken:       csrfToken,
		Flash:           flash,
		Authenticated:   authenticated,
		Title:           title,
		PageDescription: description,
	}
}

func (app *application) Serve(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     app.errLog,
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	app.infoLog.Println("listening on", addr)
	return srv.ListenAndServe()
}
