// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"html/template"
	"log"
	"net/http"

	"git.sr.ht/~kota/kudoer/models"
	"git.sr.ht/~kota/kudoer/ui"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

type application struct {
	infoLog        *log.Logger
	errLog         *log.Logger
	staticFiles    ui.StaticFiles
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
	staticFiles ui.StaticFiles,
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
		staticFiles:    staticFiles,
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
	StaticFiles     ui.StaticFiles
	CSPNonce        string
	CSRFToken       string
	Flash           string
	Authenticated   string
	Title           string
	PageDescription string
}

func (app *application) newPage(r *http.Request, title, description string) Page {
	staticFiles := app.staticFiles
	cspNonce := nonce(r.Context())
	csrfToken := nosurf.Token(r)
	flash := app.sessionManager.PopString(r.Context(), "flash")
	authenticated := app.authenticated(r)
	return Page{
		StaticFiles:     staticFiles,
		CSPNonce:        cspNonce,
		CSRFToken:       csrfToken,
		Flash:           flash,
		Authenticated:   authenticated,
		Title:           title,
		PageDescription: description,
	}
}
