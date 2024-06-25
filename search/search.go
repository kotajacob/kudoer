// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package search

import (
	"context"
	"fmt"

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
	Displayname string
}

// Open loads or creates the items and users search indexes.
// If either is missing, the entire database is used to build the missing index.
func Open(
	itemIndexPath string,
	userIndexPath string,
	items *models.ItemModel,
	users *models.UserModel,
) (itemIndex bleve.Index, userIndex bleve.Index, err error) {
	itemIndex, err = bleve.Open(itemIndexPath)
	if err != nil {
		itemMapping := bleve.NewIndexMapping()
		itemIndex, err = bleve.New(itemIndexPath, itemMapping)
		if err != nil {
			return nil, nil, err
		}

		itemIndex, err = IndexAllItems(itemIndex, items)
		if err != nil {
			return nil, nil, err
		}
	}

	userIndex, err = bleve.Open(userIndexPath)
	if err != nil {
		userMapping := bleve.NewIndexMapping()
		userIndex, err = bleve.New(userIndexPath, userMapping)
		if err != nil {
			return nil, nil, err
		}

		userIndex, err = IndexAllUsers(userIndex, users)
		if err != nil {
			return nil, nil, err
		}
	}
	return itemIndex, userIndex, nil

}

func IndexAllItems(index bleve.Index, items *models.ItemModel) (bleve.Index, error) {
	all, err := items.Index(context.Background())
	if err != nil {
		return index, err
	}

	batch := index.NewBatch()
	for _, item := range all {
		batch.Index(item.ID.String(), Item{
			ID:          item.ID.String(),
			Name:        item.Name,
			Description: item.Description,
		})
	}
	err = index.Batch(batch)
	return index, err
}

func IndexAllUsers(index bleve.Index, users *models.UserModel) (bleve.Index, error) {
	all, err := users.Index(context.Background())
	if err != nil {
		return index, err
	}

	batch := index.NewBatch()
	for _, user := range all {
		fmt.Println(user.Username) // DEBUG
		batch.Index(user.Username, User{
			Username:    user.Username,
			Displayname: user.DisplayName,
		})
	}
	err = index.Batch(batch)
	return index, err
}
