// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package db

import (
	"context"
	"embed"
	"io/fs"
	"path/filepath"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitemigration"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed "migrations"
var migrationFiles embed.FS

func Open(dsn string) (*sqlitex.Pool, error) {
	var schema sqlitemigration.Schema
	files, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		migration, err := fs.ReadFile(
			migrationFiles,
			filepath.Join("migrations", file.Name()),
		)
		if err != nil {
			return nil, err
		}

		// Each element of the Migrations slice is applied in sequence. When you
		// want to change the schema, add a new SQL script to this list.
		//
		// Existing databases will pick up at the same position in the
		// Migrations slice as they last left off.
		schema.Migrations = append(schema.Migrations, string(migration))
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
