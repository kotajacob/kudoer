// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"zombiezen.com/go/sqlite/sqlitex"
)

type application struct {
	infoLog *log.Logger
	errLog  *log.Logger

	dbpool *sqlitex.Pool
}

func main() {
	addr := flag.String("addr", ":2024", "HTTP Network Address")
	dsn := flag.String("dsn", "file:memory:?mode=memory", "SQLite data source name")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)

	dbpool, err := openDB(*dsn)
	if err != nil {
		log.Fatal(err)
	}

	app := &application{
		infoLog: infoLog,
		errLog:  errLog,
		dbpool:  dbpool,
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
	dbpool, err := sqlitex.NewPool(dsn, sqlitex.PoolOptions{
		PoolSize: 10,
	})
	if err != nil {
		return nil, err
	}
	return dbpool, nil
}
