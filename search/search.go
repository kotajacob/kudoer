// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package search

import (
	"context"
	"fmt"
	"log"
	"time"

	"git.sr.ht/~kota/kudoer/models"
	"github.com/blevesearch/bleve"
)

type Item struct {
	ID          string
	Name        string
	Description string
}

type User struct {
	Username    string
	DisplayName string
}

// Open loads or creates the items and users search indexes.
// If either is missing, the entire database is used to build the missing index.
func Open(
	infoLog *log.Logger,
	itemIndexPath string,
	userIndexPath string,
	items *models.ItemModel,
	users *models.UserModel,
) (itemIndex bleve.Index, userIndex bleve.Index, err error) {
	itemIndex, err = bleve.Open(itemIndexPath)
	if err != nil {
		itemMapping := bleve.NewIndexMapping()
		itemMapping.DefaultAnalyzer = "en"
		itemIndex, err = bleve.New(itemIndexPath, itemMapping)
		if err != nil {
			return nil, nil, err
		}

		itemIndex, err = IndexAllItems(infoLog, itemIndex, items)
		if err != nil {
			return nil, nil, fmt.Errorf("failed indexing all items: %v", err)
		}
	}

	userIndex, err = bleve.Open(userIndexPath)
	if err != nil {
		userMapping := bleve.NewIndexMapping()
		userMapping.DefaultAnalyzer = "en"
		userIndex, err = bleve.New(userIndexPath, userMapping)
		if err != nil {
			return nil, nil, err
		}

		userIndex, err = IndexAllUsers(infoLog, userIndex, users)
		if err != nil {
			return nil, nil, fmt.Errorf("failed indexing all users: %v", err)
		}
	}
	return itemIndex, userIndex, nil

}

func IndexAllItems(
	infoLog *log.Logger,
	index bleve.Index,
	items *models.ItemModel,
) (bleve.Index, error) {
	infoLog.Println("indexing all items")
	started := time.Now()

	limit := 500
	var i int
	for {
		offset := i * limit
		items, err := items.Index(context.Background(), limit, offset)
		if err != nil {
			return index, err
		}
		if len(items) == 0 {
			break
		}

		batch := index.NewBatch()
		for _, item := range items {
			batch.Index(item.ID.String(), Item{
				ID:          item.ID.String(),
				Name:        item.Name,
				Description: item.Description,
			})
		}
		err = index.Batch(batch)
		if err != nil {
			return index, err
		}
		i++
		since := time.Since(started).Round(time.Second)
		infoLog.Printf("%07d items - %v\n", offset+limit, since)
	}
	return index, nil
}

func IndexAllUsers(
	infoLog *log.Logger,
	index bleve.Index,
	users *models.UserModel,
) (bleve.Index, error) {
	infoLog.Println("indexing all users")
	started := time.Now()

	limit := 500
	var i int
	for {
		offset := i * limit
		users, err := users.Index(context.Background(), limit, offset)
		if err != nil {
			return index, err
		}
		if len(users) == 0 {
			break
		}

		batch := index.NewBatch()
		for _, user := range users {
			batch.Index(user.Username, User{
				Username:    user.Username,
				DisplayName: user.DisplayName,
			})
		}
		err = index.Batch(batch)
		if err != nil {
			return index, err
		}
		i++
		since := time.Since(started).Round(time.Second)
		infoLog.Printf("%07d users - %v\n", offset+limit, since)
	}
	return index, nil
}
