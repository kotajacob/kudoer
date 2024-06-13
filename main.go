// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"git.sr.ht/~kota/kudoer/application"
	"git.sr.ht/~kota/kudoer/litesession"
	"git.sr.ht/~kota/kudoer/models"
	"git.sr.ht/~kota/kudoer/ui"
	"github.com/alexedwards/scs/v2"
)

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
	sessionManager.ErrorFunc = func(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	) {
		trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
		errLog.Output(2, trace)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	app := application.New(
		infoLog,
		errLog,
		templates,
		sessionManager,
		&models.UserModel{DB: db},
		&models.ItemModel{DB: db},
		&models.SearchModel{DB: db},
	)

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errLog,
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("listening on", *addr)
	err = srv.ListenAndServe()
	errLog.Fatalln(err)
}
