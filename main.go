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

	"git.sr.ht/~kota/kudoer/models"
	"git.sr.ht/~kota/kudoer/ui"

	"zombiezen.com/go/sqlite/sqlitex"
)

type application struct {
	infoLog   *log.Logger
	errLog    *log.Logger
	templates map[string]*template.Template

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

	app := &application{
		infoLog:   infoLog,
		errLog:    errLog,
		templates: templates,
		kudos:     &models.KudoModel{DB: db},
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

	err = sqlitex.Execute(
		conn,
		`CREATE TABLE IF NOT EXISTS kudos (
			id TEXT NOT NULL PRIMARY KEY,
			author INTEGER NOT NULL,
			rating TEXT NOT NULL,
			body TEXT
		) WITHOUT ROWID;`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
