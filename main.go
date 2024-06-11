// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"context"
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

	"zombiezen.com/go/sqlite/sqlitex"
)

type application struct {
	infoLog        *log.Logger
	errLog         *log.Logger
	templates      map[string]*template.Template
	sessionManager *scs.SessionManager

	users *models.UserModel
	items *models.ItemModel
	kudos *models.KudoModel
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
		kudos:          &models.KudoModel{DB: db},
	}

	app.sessionManager.ErrorFunc = func(w http.ResponseWriter, r *http.Request, err error) {
		app.serverError(w, err)
		return
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errLog,
		Handler:  app.routes(),
	}

	log.Println("listening on", *addr)
	err = srv.ListenAndServe()
	errLog.Fatalln(err)
}

func openDB(dsn string) (*sqlitex.Pool, error) {
	db, err := sqlitex.NewPool(dsn, sqlitex.PoolOptions{
		PoolSize: 10,
	})
	if err != nil {
		return nil, err
	}

	conn, err := db.Take(context.Background())
	if err != nil {
		return nil, err
	}
	defer db.Put(conn)

	// Create users table.
	err = sqlitex.Execute(
		conn,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL
		) WITHOUT ROWID;`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Create users index.
	err = sqlitex.Execute(
		conn,
		`CREATE UNIQUE INDEX IF NOT EXISTS users_idx ON users (id);`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Create items table.
	err = sqlitex.Execute(
		conn,
		`CREATE TABLE IF NOT EXISTS items (
			id TEXT NOT NULL PRIMARY KEY,
			creator_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			image TEXT NOT NULL,
			FOREIGN KEY (creator_id) REFERENCES users (id)
		) WITHOUT ROWID;`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Create items index.
	err = sqlitex.Execute(
		conn,
		`CREATE UNIQUE INDEX IF NOT EXISTS items_idx ON items (id);`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Create kudos table.
	err = sqlitex.Execute(
		conn,
		`CREATE TABLE IF NOT EXISTS kudos (
			id TEXT NOT NULL PRIMARY KEY,
			item_id TEXT NOT NULL,
			creator_id TEXT NOT NULL,
			rating TEXT NOT NULL,
			body TEXT NOT NULL,
			FOREIGN KEY (creator_id) REFERENCES users (id),
			FOREIGN KEY (item_id) REFERENCES items (id)
		) WITHOUT ROWID;`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Create kudos index.
	err = sqlitex.Execute(
		conn,
		`CREATE UNIQUE INDEX IF NOT EXISTS kudos_idx ON kudos (id);`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Create sessions table.
	err = sqlitex.Execute(
		conn,
		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			data BLOB NOT NULL,
			expiry REAL NOT NULL
		);`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Create sessions index.
	err = sqlitex.Execute(
		conn,
		`CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);`,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}
