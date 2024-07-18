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
	ID   ulid.ULID
	Name string
}

type SearchUser struct {
	Username    string
	DisplayName string
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
	err = sqlitex.Execute(conn,
		`SELECT id, name FROM items_search WHERE items_search MATCH ?
		ORDER BY bm25(items_search, 0, 1) LIMIT 100`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				id, err := ulid.Parse(stmt.ColumnText(0))
				if err != nil {
					return err
				}
				items = append(items, SearchItem{
					ID:   id,
					Name: stmt.ColumnText(1),
				})
				return nil
			},
			Args: []any{query},
		})
	return items, err
}

func (m *SearchModel) Users(ctx context.Context, query string) ([]SearchUser, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return []SearchUser{}, err
	}
	defer m.DB.Put(conn)

	var users []SearchUser
	err = sqlitex.Execute(conn,
		`SELECT id, name FROM users_search WHERE users_search MATCH ?
		ORDER BY bm25(users_search, 0, 1) LIMIT 100`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				users = append(users, SearchUser{
					Username:    stmt.ColumnText(0),
					DisplayName: stmt.ColumnText(1),
				})
				return nil
			},
			Args: []any{query},
		})
	return users, err
}
