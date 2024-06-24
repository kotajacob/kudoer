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
	Username    string
	DisplayName string
	Email       string
}

type UserModel struct {
	DB *sqlitex.Pool
}

// DisplayName returns a user's display name.
func (m *UserModel) DisplayName(ctx context.Context, username string) (string, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return "", err
	}
	defer m.DB.Put(conn)

	var found bool
	var displayname string
	err = sqlitex.Execute(conn, `SELECT displayname from users WHERE username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				found = true
				displayname = stmt.ColumnText(0)
				return nil
			},
			Args: []any{username},
		})

	if !found {
		return "", ErrNoRecord
	}
	return displayname, err
}

// Get returns information about a given user.
func (m *UserModel) Get(ctx context.Context, username string) (User, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return User{}, err
	}
	defer m.DB.Put(conn)

	var u User
	err = sqlitex.Execute(conn, `SELECT displayname, email from users WHERE username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				u.Username = username

				u.DisplayName = stmt.ColumnText(0)
				u.Email = stmt.ColumnText(1)
				return nil
			},
			Args: []any{username},
		})

	if u.Username == "" {
		return u, ErrNoRecord
	}
	return u, err
}

// Insert adds a new user to the database.
func (m *UserModel) Insert(
	ctx context.Context,
	username string,
	displayname string,
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
		`INSERT INTO users (username, displayname, email, password) VALUES (?, ?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{username, displayname, email, hashedPassword}},
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

// Update a user's profile information in the database.
// Not for changing the user's password. Use ChangePassword for that.
func (m *UserModel) Update(
	ctx context.Context,
	username string,
	displayname string,
	email string,
) error {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`UPDATE users SET displayname = ?, email = ? WHERE username = ?`,
		&sqlitex.ExecOptions{Args: []any{displayname, email, username}},
	)
	return err
}

// Change a user's password.
func (m *UserModel) ChangePassword(
	ctx context.Context,
	username string,
	hashedPassword string,
) error {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`UPDATE users SET password = ? WHERE username = ?`,
		&sqlitex.ExecOptions{Args: []any{hashedPassword, username}},
	)
	return err
}

// Authenticate checks if a given username and password are correct for the
// user.
// Success is indicated with a nil error.
// Failure is indicated with ErrInvalidCredentials. All other errors are server
// errors.
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

	var found bool
	var hashedPassword string
	err = sqlitex.Execute(conn, `SELECT password from users WHERE username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				found = true
				hashedPassword = stmt.ColumnText(0)
				return nil
			},
			Args: []any{username},
		})

	if !found {
		return ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
	}
	return err
}
