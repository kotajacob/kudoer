// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package models

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Kudo struct {
	ID              ulid.ULID
	ItemID          ulid.ULID
	CreatorUsername string
	Emoji           int
	Body            string
}

type KudoModel struct {
	DB *sqlitex.Pool
}

func (m *KudoModel) Item(ctx context.Context, itemID ulid.ULID) ([]Kudo, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return []Kudo{}, err
	}
	defer m.DB.Put(conn)

	var kudos []Kudo
	err = sqlitex.Execute(conn, `SELECT id, creator_username, emoji, body from kudos WHERE item_id = ?`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			var k Kudo

			id := stmt.ColumnText(0)
			k.ID, err = ulid.Parse(id)
			if err != nil {
				return err
			}

			k.ItemID = itemID

			k.CreatorUsername = stmt.ColumnText(1)
			k.Emoji = stmt.ColumnInt(2)
			k.Body = stmt.ColumnText(3)

			kudos = append(kudos, k)

			return nil
		},
		Args: []any{itemID},
	})
	fmt.Println(kudos)
	return kudos, err
}

func (m *KudoModel) Get(ctx context.Context, id ulid.ULID) (Kudo, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return Kudo{}, err
	}
	defer m.DB.Put(conn)

	var k Kudo
	err = sqlitex.Execute(conn, `SELECT item_id, creator_username, emoji, body from kudos WHERE id = ?`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			k.ID = id

			itemID := stmt.ColumnText(0)
			k.ItemID, err = ulid.Parse(itemID)
			if err != nil {
				return err
			}

			k.CreatorUsername = stmt.ColumnText(1)
			k.Emoji = stmt.ColumnInt(2)
			k.Body = stmt.ColumnText(3)
			return nil
		},
		Args: []any{id},
	})

	if k.ID.Compare(id) != 0 {
		return k, ErrNoRecord
	}
	return k, err
}

func (m *KudoModel) Insert(
	ctx context.Context,
	item_id ulid.ULID,
	creator_username string,
	emoji int,
	body string,
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
		`INSERT INTO kudos (id, item_id, creator_username, emoji, body) VALUES (?, ?, ?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{uuid, item_id, creator_username, emoji, body}},
	)
	return uuid, err
}
