// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package models

import (
	"context"

	"github.com/oklog/ulid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type SearchItem struct {
	ID          ulid.ULID
	Name        string
	Description string
}

type SearchModel struct {
	DB *sqlitex.Pool
}

func (m *SearchModel) Items(ctx context.Context, query string) ([]SearchItem, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return []SearchItem{}, err
	}
	defer m.DB.Put(conn)

	var items []SearchItem
	err = sqlitex.Execute(conn, `SELECT id, name, description FROM items_search (?) ORDER BY rank`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			id, err := ulid.Parse(stmt.ColumnText(0))
			if err != nil {
				return err
			}
			items = append(items, SearchItem{
				ID:          id,
				Name:        stmt.ColumnText(1),
				Description: stmt.ColumnText(2),
			})
			return nil
		},
		Args: []any{query},
	})
	return items, err
}
