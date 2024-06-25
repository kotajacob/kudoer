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
	Bio         string
}

type UserModel struct {
	DB *sqlitex.Pool
}

// Index returns all users to build the initial search index.
// TODO: Support pagination.
func (m *UserModel) Index(ctx context.Context) ([]User, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return nil, err
	}
	defer m.DB.Put(conn)

	var users []User
	err = sqlitex.Execute(conn,
		`SELECT username, displayname FROM users`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				var user User
				user.Username = stmt.ColumnText(0)
				user.DisplayName = stmt.ColumnText(1)
				users = append(users, user)
				return nil
			},
		})
	return users, err
}

// Get returns information about a given user.
func (m *UserModel) Get(ctx context.Context, username string) (User, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return User{}, err
	}
	defer m.DB.Put(conn)

	var u User
	err = sqlitex.Execute(conn, `SELECT displayname, email, bio from users WHERE username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				u.Username = username

				u.DisplayName = stmt.ColumnText(0)
				u.Email = stmt.ColumnText(1)
				u.Bio = stmt.ColumnText(2)
				return nil
			},
			Args: []any{username},
		})

	if u.Username == "" {
		return u, ErrNoRecord
	}
	return u, err
}

type SortedUsernames []string

// GetList returns information for each user in a list of usernames.
// The index of the given users is used to sort the result.
// That way you can get your list back in the same order you gave it in.
func (m *UserModel) GetList(
	ctx context.Context,
	usernames SortedUsernames,
) ([]User, error) {
	// Early exit if given an empty list!
	if len(usernames) == 0 {
		return []User{}, nil
	}

	conn, err := m.DB.Take(ctx)
	if err != nil {
		return nil, err
	}
	defer m.DB.Put(conn)

	// Create a temporary table to store the sortedUsernames.
	err = sqlitex.Execute(conn, `CREATE TEMP TABLE sorted_usernames (
		idx INTEGER NOT NULL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE
	);`, nil)
	if err != nil {
		return nil, err
	}

	// Fill the temporary table.
	var q strings.Builder
	var args []any
	q.WriteString(`INSERT INTO sorted_usernames (idx, username) VALUES `)
	for i, username := range usernames {
		if i != 0 {
			q.WriteString(`,`)
		}
		q.WriteString(`(?, ?)`)
		args = append(args, i)
		args = append(args, username)
	}
	q.WriteString(`;`)
	err = sqlitex.Execute(conn,
		q.String(),
		&sqlitex.ExecOptions{
			Args: args,
		})
	if err != nil {
		return nil, err
	}

	// Join the temporary table and the users table using the temporary tables
	// index to sort the output.
	var users []User
	err = sqlitex.Execute(conn,
		`SELECT users.username, users.displayname
		FROM temp.sorted_usernames JOIN users ON temp.sorted_usernames.username = users.username
		ORDER BY temp.sorted_usernames.idx;`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				var u User
				u.Username = stmt.ColumnText(0)
				u.DisplayName = stmt.ColumnText(1)
				users = append(users, u)
				return nil
			},
		})
	return users, err
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
	bio string,
) error {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`UPDATE users SET displayname = ?, email = ?, bio = ? WHERE username = ?`,
		&sqlitex.ExecOptions{Args: []any{displayname, email, bio, username}},
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
