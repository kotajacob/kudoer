// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"context"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitemigration"
	"zombiezen.com/go/sqlite/sqlitex"
)

func openDB(dsn string) (*sqlitex.Pool, error) {
	schema := sqlitemigration.Schema{
		// Each element of the Migrations slice is applied in sequence. When you
		// want to change the schema, add a new SQL script to this list.
		//
		// Existing databases will pick up at the same position in the Migrations
		// slice as they last left off.
		Migrations: []string{
			`CREATE TABLE IF NOT EXISTS users (
			username TEXT NOT NULL PRIMARY KEY,
			displayname TEXT NOT NULL,
			email TEXT NOT NULL,
			password TEXT NOT NULL
		) WITHOUT ROWID;`,
			`CREATE UNIQUE INDEX IF NOT EXISTS users_idx ON users (username);`,

			`CREATE TABLE IF NOT EXISTS items (
			id TEXT NOT NULL PRIMARY KEY,
			creator_username TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			image TEXT NOT NULL,
			FOREIGN KEY (creator_username) REFERENCES users (username)
		) WITHOUT ROWID;`,
			`CREATE UNIQUE INDEX IF NOT EXISTS items_idx ON items (id);`,

			`CREATE VIRTUAL TABLE IF NOT EXISTS items_search USING fts5(
			id,
			name,
			description
		);`,
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

			`CREATE TABLE IF NOT EXISTS kudos (
			id TEXT NOT NULL PRIMARY KEY,
			item_id TEXT NOT NULL,
			creator_username TEXT NOT NULL,
			emoji INTEGER NOT NULL,
			body TEXT NOT NULL,
			FOREIGN KEY (creator_username) REFERENCES users (username)
		) WITHOUT ROWID;`,
			`CREATE UNIQUE INDEX IF NOT EXISTS kudos_idx ON kudos (id);`,
			`CREATE INDEX IF NOT EXISTS kudos_item_idx ON kudos (item_id);`,
			`CREATE INDEX IF NOT EXISTS kudos_creator_usernamex ON kudos (creator_username);`,

			`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			data BLOB NOT NULL,
			expiry REAL NOT NULL
		);`,
			`CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);`,
		},
	}

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

	// Run migrations.
	conn, err := db.Take(context.Background())
	if err != nil {
		return nil, err
	}
	defer db.Put(conn)
	err = sqlitemigration.Migrate(context.Background(), conn, schema)
	if err != nil {
		return nil, err
	}
	return db, nil
}
