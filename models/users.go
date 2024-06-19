// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package models

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
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

	var u User
	err = sqlitex.Execute(conn, `SELECT email from users WHERE username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				u.Username = username

				u.Email = stmt.ColumnText(0)
				return nil
			},
			Args: []any{username},
		})

	if u.Username == "" {
		return u, ErrNoRecord
	}
	return u, err
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

func (m *UserModel) Authenticate(
	ctx context.Context,
	username string,
	password string,
) error {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return err
	}
	defer m.DB.Put(conn)

	var u User
	err = sqlitex.Execute(conn, `SELECT password from users WHERE username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				u.Username = username

				u.HashedPassword = stmt.ColumnText(0)
				return nil
			},
			Args: []any{username},
		})

	if u.Username == "" {
		return ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
	}
	return err
}
