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

type User struct {
	ID       ulid.ULID
	Username string
	Email    string
}

type UserModel struct {
	DB *sqlitex.Pool
}

func (m *UserModel) Get(ctx context.Context, uuid ulid.ULID) (User, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return User{}, err
	}
	defer m.DB.Put(conn)

	var k User
	err = sqlitex.Execute(conn, `SELECT username, email from users WHERE id = ?`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			k.ID = uuid
			k.Username = stmt.ColumnText(0)
			k.Email = stmt.ColumnText(1)
			return nil
		},
		Args: []any{uuid},
	})

	if k.ID.Compare(uuid) != 0 {
		return k, ErrNoRecord
	}
	return k, err
}

func (m *UserModel) Insert(
	ctx context.Context,
	username string,
	email string,
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
		`INSERT INTO users (id, username, email) VALUES (?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{uuid, username, email}},
	)
	return uuid, err
}
