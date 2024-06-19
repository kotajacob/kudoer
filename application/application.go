// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"html/template"
	"log"
	"net/http"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/alexedwards/scs/v2"
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

type Page struct {
	CSPNonce      string
	Flash         string
	Authenticated string
}

func (app *application) newPage(r *http.Request) Page {
	cspNonce := nonce(r.Context())
	flash := app.sessionManager.PopString(r.Context(), "flash")
	authenticated := app.authenticated(r)
	return Page{
		CSPNonce:      cspNonce,
		Flash:         flash,
		Authenticated: authenticated,
	}
}
