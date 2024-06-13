// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"context"
	"html/template"
	"log"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/alexedwards/scs/v2"
)

type application struct {
	infoLog        *log.Logger
	errLog         *log.Logger
	templates      map[string]*template.Template
	sessionManager *scs.SessionManager

	users *models.UserModel
	items *models.ItemModel
}

func New(infoLog *log.Logger, errLog *log.Logger, templates map[string]*template.Template, sessionManager *scs.SessionManager, users *models.UserModel, items *models.ItemModel) *application {
	return &application{
		infoLog:        infoLog,
		errLog:         errLog,
		templates:      templates,
		sessionManager: sessionManager,
		users:          users,
		items:          items,
	}
}

type Page struct {
	CSPNonce string
	Flash    string
}

func (app *application) newPage(ctx context.Context) Page {
	cspNonce := nonce(ctx)
	flash := app.sessionManager.PopString(ctx, "flash")
	return Page{
		CSPNonce: cspNonce,
		Flash:    flash,
	}
}
