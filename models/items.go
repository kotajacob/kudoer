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
	ID          ulid.ULID
	CreatorID   ulid.ULID
	Name        string
	Description string
	Image       string
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

	var k Item
	err = sqlitex.Execute(conn, `SELECT creator_id, name, description, image from items WHERE id = ?`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			k.ID = uuid

			creator_id, err := ulid.Parse(stmt.ColumnText(0))
			if err != nil {
				return err
			}
			k.CreatorID = creator_id
			k.Name = stmt.ColumnText(1)
			k.Description = stmt.ColumnText(2)
			k.Image = stmt.ColumnText(3)
			return nil
		},
		Args: []any{uuid},
	})

	if k.ID.Compare(uuid) != 0 {
		return k, ErrNoRecord
	}
	return k, err
}

func (m *ItemModel) Insert(
	ctx context.Context,
	creator_id ulid.ULID,
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
		`INSERT INTO items (id, creator_id, name, description, image) VALUES (?, ?, ?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{uuid, creator_id, name, description, image}},
	)
	return uuid, err
}
