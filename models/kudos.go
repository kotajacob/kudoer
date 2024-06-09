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
	ID     ulid.ULID
	Author int
	Rating string // emoji,
	Body   string
}

type KudoModel struct {
	DB *sqlitex.Pool
}

func (m *KudoModel) Get(ctx context.Context, uuid ulid.ULID) (Kudo, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return Kudo{}, err
	}
	defer m.DB.Put(conn)

	var k Kudo
	err = sqlitex.Execute(conn, `SELECT author, rating, body from kudos WHERE id = ?`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			k.ID = uuid
			k.Author = stmt.ColumnInt(0)
			k.Rating = stmt.ColumnText(1)
			k.Body = stmt.ColumnText(2)
			return nil
		},
		Args: []any{uuid},
	})

	if k.ID.Compare(uuid) != 0 {
		return k, ErrNoRecord
	}
	return k, err
}

func (m *KudoModel) Insert(
	ctx context.Context,
	author int,
	rating,
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
		`INSERT INTO kudos (id, author, rating, body) VALUES (?, ?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{uuid, author, rating, body}},
	)
	return uuid, err
}
