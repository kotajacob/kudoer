package models

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/oklog/ulid"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Kudo struct {
	ID     int
	Author int
	Rating string // emoji,
	Body   string
}

type KudoModel struct {
	DB *sqlitex.Pool
}

func (m *KudoModel) Get(id ulid.ULID) (Kudo, error) {
	return Kudo{}, nil
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
	if err != nil {
		return uuid, err
	}
	return uuid, nil
}
