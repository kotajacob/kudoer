// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package models

import (
	"context"
	"strings"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type User struct {
	Username       string
	Email          string
	HashedPassword string
}

type UserModel struct {
	DB *sqlitex.Pool
}

func (m *UserModel) Get(ctx context.Context, username string) (User, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return User{}, err
	}
	defer m.DB.Put(conn)

	var k User
	err = sqlitex.Execute(conn, `SELECT email from users WHERE username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				k.Username = username

				k.Email = stmt.ColumnText(0)
				return nil
			},
			Args: []any{username},
		})

	if k.Username == "" {
		return k, ErrNoRecord
	}
	return k, err
}

func (m *UserModel) Insert(
	ctx context.Context,
	username string,
	email string,
	hashedPassword string,
) error {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`INSERT INTO users (username, email, password) VALUES (?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{username, email, hashedPassword}},
	)
	if sqlite.ErrCode(err) == sqlite.ResultConstraintUnique {
		if strings.HasSuffix(err.Error(), "users.username") {
			return ErrUsernameExists
		}
		if strings.HasSuffix(err.Error(), "users.email") {
			return ErrEmailExists
		}
		return err
	}
	return err
}
