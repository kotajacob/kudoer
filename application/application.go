// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"bytes"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"html/template"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"git.sr.ht/~kota/kudoer/application/mail"
	"git.sr.ht/~kota/kudoer/application/media"
	"git.sr.ht/~kota/kudoer/db/models"
	"github.com/alexedwards/scs/v2"
	"github.com/disintegration/imaging"
	"github.com/justinas/nosurf"
	"github.com/throttled/throttled/v2"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type application struct {
	infoLog        *log.Logger
	errLog         *log.Logger
	templates      map[string]*template.Template
	sessionManager *scs.SessionManager
	rateLimiter    *throttled.HTTPRateLimiterCtx
	mediaStore     *media.MediaStore
	mailer         *mail.Mailer

	users       *models.UserModel
	items       *models.ItemModel
	kudos       *models.KudoModel
	search      *models.SearchModel
	pwresets    *models.PWResetModel
	profilepics *models.ProfilePictureModel
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
	profilepics *models.ProfilePictureModel,
) *application {
	app := &application{
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
		profilepics:    profilepics,
	}

	// TODO: Remove after migration is completed!
	app.MigrateProfilePics()

	return app
}

// TODO: Remove after migration is completed!
func (app *application) MigrateProfilePics() {
	app.infoLog.Println("running profile picture migration")

	conn, err := app.users.DB.Take(context.Background())
	if err != nil {
		app.errLog.Fatalln(err)
	}
	defer app.users.DB.Put(conn)

	type user struct {
		username string
		pic      string
	}
	var users []user
	err = sqlitex.Execute(
		conn,
		`SELECT username, pic FROM users`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				var u user
				u.username = stmt.ColumnText(0)
				u.pic = stmt.ColumnText(1)

				users = append(users, u)
				return nil
			},
		},
	)
	if err != nil {
		app.errLog.Fatalln(err)
	}

	for _, u := range users {
		// See if there's a matching file.
		if u.pic == "" {
			app.infoLog.Println(u.username, "skipping: has no pic")
			continue
		}
		_, err := os.Stat(filepath.Join(app.mediaStore.Dir(), u.pic))
		if err != nil {
			app.infoLog.Println(u.username, "skipping: file does not exist")
			continue
		}
		app.infoLog.Println(u.username, "migrating")

		// Set 512 variant.
		err = sqlitex.Execute(
			conn,
			`INSERT INTO profile_pictures
			(filename, username, kind) VALUES (?, ?, ?)`,
			&sqlitex.ExecOptions{Args: []any{u.pic, u.username, models.ProfileJPEG512}},
		)
		if err != nil {
			app.errLog.Fatalln(err)
		}

		// Create 128 variant.
		app.infoLog.Println("opening", filepath.Join(app.mediaStore.Dir(), u.pic))
		img, err := imaging.Open(filepath.Join(app.mediaStore.Dir(), u.pic))
		if err != nil {
			app.errLog.Fatalln(err)
		}
		// Scale / shrink to the correct final size.
		img = imaging.Fill(
			img,
			128,
			128,
			imaging.Center,
			imaging.Lanczos,
		)

		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
		if err != nil {
			app.errLog.Fatalln(err)
		}

		// Calculate hash for the filename.
		h := sha1.New()
		r := bytes.NewReader(buf.Bytes())
		if _, err := io.Copy(h, r); err != nil {
			app.errLog.Fatalln(err)
		}
		r.Seek(0, 0)

		name := fmt.Sprintf("%x.jpeg", h.Sum(nil))
		f, err := os.Create(filepath.Join(app.mediaStore.Dir(), name))
		if err != nil {
			app.errLog.Fatalln(err)
		}

		if _, err := io.Copy(f, r); err != nil {
			app.errLog.Fatalln(err)
		}
		err = f.Close()
		if err != nil {
			app.errLog.Fatalln(err)
		}
		err = sqlitex.Execute(
			conn,
			`INSERT INTO profile_pictures
			(filename, username, kind) VALUES (?, ?, ?)`,
			&sqlitex.ExecOptions{Args: []any{name, u.username, models.ProfileJPEG128}},
		)
		if err != nil {
			app.errLog.Fatalln(err)
		}
	}
	app.infoLog.Println("finished migrations")
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
