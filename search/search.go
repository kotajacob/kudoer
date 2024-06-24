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

func Open(path string, items *models.ItemModel) (bleve.Index, error) {
	index, err := bleve.Open(path)
	if err == nil {
		return index, nil
	}

	mapping := bleve.NewIndexMapping()
	index, err = bleve.New(path, mapping)
	if err != nil {
		return nil, err
	}

	return IndexAll(index, items)
}

func IndexAll(index bleve.Index, items *models.ItemModel) (bleve.Index, error) {
	all, err := items.Index(context.Background())
	if err != nil {
		return index, err
	}

	batch := index.NewBatch()
	for _, item := range all {
		fmt.Println(item.Name)
		batch.Index(item.ID.String(), Item{
			ID:          item.ID.String(),
			Name:        item.Name,
			Description: item.Description,
		})
	}
	err = index.Batch(batch)
	return index, err
}
