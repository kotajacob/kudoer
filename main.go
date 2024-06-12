// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"git.sr.ht/~kota/kudoer/litesession"
	"git.sr.ht/~kota/kudoer/models"
	"git.sr.ht/~kota/kudoer/ui"
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

func main() {
	addr := flag.String("addr", ":2024", "HTTP Network Address")
	dsn := flag.String("dsn", "kudoer.db", "SQLite data source name")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	templates, err := ui.Templates()
	if err != nil {
		errLog.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Store = litesession.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		infoLog:        infoLog,
		errLog:         errLog,
		templates:      templates,
		sessionManager: sessionManager,
		users:          &models.UserModel{DB: db},
		items:          &models.ItemModel{DB: db},
	}

	app.sessionManager.ErrorFunc = func(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	) {
		app.serverError(w, err)
		return
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("listening on", *addr)
	err = srv.ListenAndServe()
	errLog.Fatalln(err)
}
