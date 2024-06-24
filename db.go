// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"context"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func openDB(dsn string) (*sqlitex.Pool, error) {
	db, err := sqlitex.NewPool(dsn, sqlitex.PoolOptions{
		PoolSize: 10,
		PrepareConn: func(conn *sqlite.Conn) error {
			// Create users table.
			err := sqlitex.Execute(
				conn,
				`PRAGMA foreign_keys = on`,
				nil,
			)
			return err
		},
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
			username TEXT NOT NULL PRIMARY KEY,
			displayname TEXT NOT NULL,
			email TEXT NOT NULL,
			password TEXT NOT NULL
		) WITHOUT ROWID;`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Create users index.
	err = sqlitex.Execute(
		conn,
		`CREATE UNIQUE INDEX IF NOT EXISTS users_idx ON users (username);`,
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
			creator_username TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			image TEXT NOT NULL,
			FOREIGN KEY (creator_username) REFERENCES users (username)
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

	// Create items full text search.
	err = sqlitex.Execute(
		conn,
		`CREATE VIRTUAL TABLE IF NOT EXISTS items_search USING fts5(
			id,
			name,
			description
		);`,
		nil,
	)
	if err != nil {
		return nil, err
	}
	err = sqlitex.Execute(
		conn,
		`CREATE TRIGGER IF NOT EXISTS after_items_insert AFTER INSERT ON items
			BEGIN INSERT INTO items_search(
				id,
				name,
				description
			)
			VALUES (
				new.id,
				new.name,
				new.description
			);
		END;`,
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
			creator_username TEXT NOT NULL,
			emoji INTEGER NOT NULL,
			body TEXT NOT NULL,
			FOREIGN KEY (creator_username) REFERENCES users (username)
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
	err = sqlitex.Execute(
		conn,
		`CREATE INDEX IF NOT EXISTS kudos_item_idx ON kudos (item_id);`,
		nil,
	)
	if err != nil {
		return nil, err
	}
	err = sqlitex.Execute(
		conn,
		`CREATE INDEX IF NOT EXISTS kudos_creator_usernamex ON kudos (creator_username);`,
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
