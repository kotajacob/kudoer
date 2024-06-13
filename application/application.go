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
	search *models.SearchModel
}

func New(
	infoLog *log.Logger,
	errLog *log.Logger,
	templates map[string]*template.Template,
	sessionManager *scs.SessionManager,
	users *models.UserModel,
	items *models.ItemModel,
	search *models.SearchModel,
) *application {
	return &application{
		infoLog:        infoLog,
		errLog:         errLog,
		templates:      templates,
		sessionManager: sessionManager,
		users:          users,
		items:          items,
		search:         search,
	}
}

type Page struct {
	CSPNonce        string
	Flash           string
	IsAuthenticated bool
}

func (app *application) newPage(r *http.Request) Page {
	cspNonce := nonce(r.Context())
	flash := app.sessionManager.PopString(r.Context(), "flash")
	isAuthenticated := app.isAuthenticated(r)
	return Page{
		CSPNonce:        cspNonce,
		Flash:           flash,
		IsAuthenticated: isAuthenticated,
	}
}
