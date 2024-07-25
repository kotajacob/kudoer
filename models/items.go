// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package models

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Item struct {
	ID              ulid.ULID
	CreatorUsername string
	Name            string
	Description     string
	Source          string
}

type ItemModel struct {
	DB *sqlitex.Pool
}

// Info returns information about a given item.
func (m *ItemModel) Info(ctx context.Context, uuid ulid.ULID) (Item, error) {
	conn, err := m.DB.Take(ctx)
	if err != nil {
		return Item{}, err
	}
	defer m.DB.Put(conn)

	var i Item
	err = sqlitex.Execute(conn,
		`SELECT creator_username, name, description, source FROM items WHERE id = ?`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				i.ID = uuid

				i.CreatorUsername = stmt.ColumnText(0)
				i.Name = stmt.ColumnText(1)
				i.Description = stmt.ColumnText(2)
				i.Source = stmt.ColumnText(3)
				return nil
			},
			Args: []any{uuid},
		})

	if i.ID.Compare(uuid) != 0 {
		fmt.Println("uuid", uuid.String(), i.ID.String())
		return i, ErrNoRecord
	}
	return i, err
}

type SortedIDs []string

// ListInfo returns information for each item in a list of IDs.
// The index of the given items is used to sort the result.
// That way you can get your list back in the same order you gave it in.
func (m *ItemModel) ListInfo(
	ctx context.Context,
	ids SortedIDs,
) ([]Item, error) {
	// Early exit if given an empty list!
	if len(ids) == 0 {
		return []Item{}, nil
	}

	conn, err := m.DB.Take(ctx)
	if err != nil {
		return nil, err
	}
	defer m.DB.Put(conn)

	// Create a temporary table to store the sortedIDs.
	err = sqlitex.Execute(conn, `CREATE TEMP TABLE sorted_ids (
		idx INTEGER NOT NULL PRIMARY KEY,
		id TEXT NOT NULL UNIQUE
	);`, nil)
	if err != nil {
		return nil, err
	}

	// Fill the temporary table.
	var q strings.Builder
	var args []any
	q.WriteString(`INSERT INTO sorted_ids (idx, id) VALUES `)
	for i, id := range ids {
		if i != 0 {
			q.WriteString(`,`)
		}
		q.WriteString(`(?, ?)`)
		args = append(args, i)
		args = append(args, id)
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

	// Join the temporary table and the items table using the temporary tables
	// index to sort the output.
	var items []Item
	err = sqlitex.Execute(conn,
		`SELECT items.id, items.creator_username, items.name, items.description
		FROM temp.sorted_ids JOIN items ON temp.sorted_ids.id = items.id
		ORDER BY temp.sorted_ids.idx;`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				var i Item
				uuid, err := ulid.Parse(stmt.ColumnText(0))
				if err != nil {
					return err
				}
				i.ID = uuid

				i.CreatorUsername = stmt.ColumnText(1)
				i.Name = stmt.ColumnText(2)
				i.Description = stmt.ColumnText(3)
				items = append(items, i)
				return nil
			},
		})
	if err != nil {
		return nil, err
	}

	err = sqlitex.Execute(conn, `DROP TABLE IF EXISTS tmp.sorted_ids`, nil)
	return items, err
}

// Insert adds a new item to the database.
func (m *ItemModel) Insert(
	ctx context.Context,
	creator_username string,
	name string,
	description string,
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
		`INSERT INTO items (id, creator_username, name, description) VALUES (?, ?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{uuid, creator_username, name, description}},
	)
	return uuid, err
}
