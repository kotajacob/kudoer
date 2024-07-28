// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.sr.ht/~kota/kudoer/application/mail"
	"git.sr.ht/~kota/kudoer/application/media"
	"git.sr.ht/~kota/kudoer/db/models"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
	"github.com/throttled/throttled/v2"
)

type application struct {
	infoLog        *log.Logger
	errLog         *log.Logger
	templates      map[string]*template.Template
	sessionManager *scs.SessionManager
	rateLimiter    *throttled.HTTPRateLimiterCtx
	mediaStore     *media.MediaStore
	mailer         *mail.Mailer

	users    *models.UserModel
	items    *models.ItemModel
	kudos    *models.KudoModel
	search   *models.SearchModel
	pwresets *models.PWResetModel
}

func New(
	infoLog *log.Logger,
	errLog *log.Logger,
	templates map[string]*template.Template,
	sessionManager *scs.SessionManager,
	rateLimiter *throttled.HTTPRateLimiterCtx,
	mediaStore *media.MediaStore,
	mailer *mail.Mailer,
	users *models.UserModel,
	items *models.ItemModel,
	kudos *models.KudoModel,
	search *models.SearchModel,
	pwresets *models.PWResetModel,
) *application {
	return &application{
		infoLog:        infoLog,
		errLog:         errLog,
		templates:      templates,
		sessionManager: sessionManager,
		rateLimiter:    rateLimiter,
		mediaStore:     mediaStore,
		mailer:         mailer,
		users:          users,
		items:          items,
		kudos:          kudos,
		search:         search,
		pwresets:       pwresets,
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

	// Handle shutdown signals gracefully.
	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.infoLog.Println("shutting down server:", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.infoLog.Println("listening on", addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.infoLog.Println("stopped server")
	return nil
}
