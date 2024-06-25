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
	ID                 ulid.ULID
	ItemID             ulid.ULID
	ItemName           string
	CreatorUsername    string
	CreatorDisplayName string
	Emoji              int
	Body               string
}

type KudoModel struct {
	DB *sqlitex.Pool
}

// Item returns all kudos for a given item.
// TODO: Add pagination
func (m *KudoModel) Item(ctx context.Context, itemID ulid.ULID) ([]Kudo, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return []Kudo{}, err
	}
	defer m.DB.Put(conn)

	var kudos []Kudo
	err = sqlitex.Execute(conn,
		`SELECT kudos.id, items.name, kudos.creator_username, users.displayname, kudos.emoji, kudos.body FROM kudos JOIN users ON kudos.creator_username = users.username JOIN items on kudos.item_id = items.id WHERE kudos.item_id = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				var k Kudo

				id := stmt.ColumnText(0)
				k.ID, err = ulid.Parse(id)
				if err != nil {
					return err
				}

				k.ItemID = itemID

				k.ItemName = stmt.ColumnText(1)
				k.CreatorUsername = stmt.ColumnText(2)
				k.CreatorDisplayName = stmt.ColumnText(3)
				k.Emoji = stmt.ColumnInt(4)
				k.Body = stmt.ColumnText(5)

				kudos = append(kudos, k)

				return nil
			},
			Args: []any{itemID},
		})
	return kudos, err
}

// ItemUser returns the kudo for a given combination of item and user if it
// exists.
// TODO: Add pagination
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
		`SELECT kudos.id, items.name, users.displayname, kudos.emoji, kudos.body FROM kudos JOIN users ON kudos.creator_username = users.username JOIN items on kudos.item_id = items.id WHERE kudos.item_id = ? AND kudos.creator_username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				id := stmt.ColumnText(0)
				k.ID, err = ulid.Parse(id)
				if err != nil {
					return err
				}

				k.ItemID = itemID
				k.ItemName = stmt.ColumnText(1)
				k.CreatorUsername = creator_username
				k.CreatorDisplayName = stmt.ColumnText(2)
				k.Emoji = stmt.ColumnInt(3)
				k.Body = stmt.ColumnText(4)

				return nil
			},
			Args: []any{itemID, creator_username},
		})

	if k.ItemID.Compare(itemID) != 0 {
		return k, ErrNoRecord
	}
	return k, err
}

// User returns all kudos for a given user.
// TODO: Add pagination
func (m *KudoModel) User(
	ctx context.Context,
	creator_username string,
) ([]Kudo, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return []Kudo{}, err
	}
	defer m.DB.Put(conn)

	var kudos []Kudo
	err = sqlitex.Execute(conn,
		`SELECT kudos.id, kudos.item_id, items.name, users.displayname, kudos.emoji, kudos.body FROM kudos JOIN users ON kudos.creator_username = users.username JOIN items on kudos.item_id = items.id WHERE kudos.creator_username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				var k Kudo

				id := stmt.ColumnText(0)
				k.ID, err = ulid.Parse(id)
				if err != nil {
					return err
				}

				itemID := stmt.ColumnText(1)
				k.ItemID, err = ulid.Parse(itemID)
				if err != nil {
					return err
				}

				k.ItemName = stmt.ColumnText(2)
				k.CreatorUsername = creator_username
				k.CreatorDisplayName = stmt.ColumnText(3)
				k.Emoji = stmt.ColumnInt(4)
				k.Body = stmt.ColumnText(5)

				kudos = append(kudos, k)
				return nil
			},
			Args: []any{creator_username},
		})
	return kudos, err
}

// Insert a kudo.
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

// Update a kudo.
func (m *KudoModel) Update(
	ctx context.Context,
	id ulid.ULID,
	item_id ulid.ULID,
	creator_username string,
	emoji int,
	body string,
) error {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`UPDATE kudos SET item_id = ?, creator_username = ?, emoji = ?, body = ? WHERE id = ?`,
		&sqlitex.ExecOptions{Args: []any{item_id, creator_username, emoji, body, id}},
	)
	return err
}
