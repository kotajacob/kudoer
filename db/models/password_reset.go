// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const pwresetTTL = 45 * time.Minute

type token struct {
	Plaintext string
	Hash      []byte
	Username  string
	Expiry    time.Time
}

// PWResetModel handles password reset request storage.
type PWResetModel struct {
	DB *sqlitex.Pool
}

// New creates a password reset token, stores the hash in the database,
// and returns the plaintext version to be sent to the user.
//
// New will delete ALL existing tokens for this username before creating a new
// token.
func (m *PWResetModel) New(
	ctx context.Context,
	username string,
) (string, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return "", err
	}
	defer m.DB.Put(conn)

	// Delete existing tokens.
	err = m.DeleteAllUser(ctx, username)
	if err != nil {
		return "", err
	}

	token, err := generatePWResetToken(username)
	if err != nil {
		return "", err
	}

	err = sqlitex.Execute(
		conn,
		`INSERT INTO pwreset_tokens (hash, username, expiry) VALUES (?, ?, ?)`,
		&sqlitex.ExecOptions{
			Args: []any{token.Hash, username, token.Expiry.Unix()},
		},
	)
	return token.Plaintext, err
}

// generatePWResetToken generates a new password reset token for a given user.
func generatePWResetToken(username string) (token, error) {
	token := token{
		Username: username,
		Expiry:   time.Now().Add(pwresetTTL),
	}
	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return token, err
	}

	// This creates a nice looking string of capital letters and numbers to
	// send to the user.
	// Padding not needed or wanted.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Password reset tokens are high-entropy (128 bits) -- unlike a random
	// user's password. As a result it's sufficient to use a faster hashing
	// algorithm rather than bcrypt.
	// https://security.stackexchange.com/questions/151257/what-kind-of-hashing-to-use-for-storing-rest-api-tokens-in-the-database
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

// Validate checks if a token exists and is still valid for a given user.
func (m *PWResetModel) Validate(
	ctx context.Context,
	token string,
) (string, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return "", err
	}
	defer m.DB.Put(conn)

	sum := sha256.Sum256([]byte(token))
	hash := sum[:]

	var username string
	var valid bool
	err = sqlitex.Execute(
		conn,
		`SELECT username FROM pwreset_tokens
		WHERE hash = ? AND expiry > ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				username = stmt.ColumnText(0)
				valid = true
				return nil
			},
			Args: []any{
				hash,
				time.Now().Unix(),
			},
		},
	)
	if err != nil {
		return "", err
	}

	if !valid {
		return "", ErrPWResetTokenInvalid
	}
	return username, nil
}

// DeleteAllUser deletes all password reset tokens for a given user.
func (m *PWResetModel) DeleteAllUser(
	ctx context.Context,
	username string,
) error {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`DELETE FROM pwreset_tokens WHERE username = ?`,
		&sqlitex.ExecOptions{
			Args: []any{username},
		},
	)
	return err
}
