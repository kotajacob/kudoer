// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package models

import (
	"context"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// ProfilePicture is a mapping of all of a user's profile picture formats to
// their filenames.
type ProfilePicture map[ProfilePictureKind]string

type ProfilePictureKind int

const (
	ProfileJPEG512 ProfilePictureKind = iota
	ProfileJPEG128
)

// ProfilePictureModel handles profile picture metadata storage.
type ProfilePictureModel struct {
	DB *sqlitex.Pool
}

// Set a user's profile picture.
// The user's old profile picture filenames, if they exist, are returned.
func (m *ProfilePictureModel) Set(
	ctx context.Context,
	username string,
	pic512 string,
	pic128 string,
) error {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`INSERT OR REPLACE INTO profile_pictures
		(filename, username, kind) VALUES (?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{pic512, username, ProfileJPEG512}},
	)
	if err != nil {
		return err
	}

	err = sqlitex.Execute(
		conn,
		`INSERT OR REPLACE INTO profile_pictures
		(filename, username, kind) VALUES (?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{pic128, username, ProfileJPEG128}},
	)
	return err
}

// Get a user's profile pictures.
func (m *ProfilePictureModel) Get(
	ctx context.Context,
	username string,
) (ProfilePicture, error) {
	pp := make(ProfilePicture)
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return pp, err
	}
	defer m.DB.Put(conn)

	err = sqlitex.Execute(
		conn,
		`SELECT filename, kind FROM profile_pictures WHERE username = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				filename := stmt.ColumnText(0)
				kind := stmt.ColumnInt(1)
				pp[ProfilePictureKind(kind)] = filename
				return nil
			},
			Args: []any{username},
		},
	)
	return pp, err
}
