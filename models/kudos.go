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

func (m *KudoModel) ItemAll(ctx context.Context, itemID ulid.ULID) ([]Kudo, error) {
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
	return kudos, err
}

func (m *KudoModel) ItemUser(
	ctx context.Context,
	itemID ulid.ULID,
	creator_username string,
) (Kudo, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return Kudo{}, err
	}
	defer m.DB.Put(conn)

	var k Kudo
	err = sqlitex.Execute(conn,
		`SELECT id, emoji, body from kudos WHERE item_id = ? AND creator_username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				id := stmt.ColumnText(0)
				k.ID, err = ulid.Parse(id)
				if err != nil {
					return err
				}

				k.ItemID = itemID
				k.CreatorUsername = creator_username

				k.Emoji = stmt.ColumnInt(1)
				k.Body = stmt.ColumnText(2)

				return nil
			},
			Args: []any{itemID, creator_username},
		})

	if k.ItemID.Compare(itemID) != 0 {
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

func (m *KudoModel) Update(
	ctx context.Context,
	id ulid.ULID,
	item_id ulid.ULID,
	creator_username string,
	emoji int,
	body string,
) (ulid.ULID, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return id, err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`UPDATE kudos SET item_id = ?, creator_username = ?, emoji = ?, body = ? WHERE id = ?`,
		&sqlitex.ExecOptions{Args: []any{item_id, creator_username, emoji, body, id}},
	)
	return id, err
}
