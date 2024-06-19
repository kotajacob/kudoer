// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package models

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/oklog/ulid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Item struct {
	ID              ulid.ULID
	CreatorUsername string
	Name            string
	Description     string
	Image           string
}

type ItemModel struct {
	DB *sqlitex.Pool
}

func (m *ItemModel) Get(ctx context.Context, uuid ulid.ULID) (Item, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return Item{}, err
	}
	defer m.DB.Put(conn)

	var i Item
	err = sqlitex.Execute(conn, `SELECT creator_username, name, description, image from items WHERE id = ?`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			i.ID = uuid

			i.CreatorUsername = stmt.ColumnText(0)
			i.Name = stmt.ColumnText(1)
			i.Description = stmt.ColumnText(2)
			i.Image = stmt.ColumnText(3)
			return nil
		},
		Args: []any{uuid},
	})

	if i.ID.Compare(uuid) != 0 {
		return i, ErrNoRecord
	}
	return i, err
}

func (m *ItemModel) Insert(
	ctx context.Context,
	creator_username string,
	name string,
	description string,
	image string,
) (ulid.ULID, error) {
	ms := ulid.Timestamp(time.Now())
	uuid, err := ulid.New(ms, rand.Reader)
	if err != nil {
		return uuid, err
	}

	conn, err := m.DB.Take(ctx)
	if err != nil {
		return uuid, err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`INSERT INTO items (id, creator_username, name, description, image) VALUES (?, ?, ?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{uuid, creator_username, name, description, image}},
	)
	return uuid, err
}
