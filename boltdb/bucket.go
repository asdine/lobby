package boltdb

import (
	"github.com/asdine/brazier"
	"github.com/asdine/brazier/boltdb/internal"
	"github.com/asdine/storm"
	"github.com/pkg/errors"
)

var _ brazier.Bucket = new(Bucket)

// NewBucket returns a Bucket
func NewBucket(node storm.Node) *Bucket {
	return &Bucket{
		node: node,
	}
}

// Bucket is a BoltDB implementation of a bucket.
type Bucket struct {
	node storm.Node
}

// Save data to the bucket. Returns an Item.
func (b *Bucket) Save(key string, data []byte) (*brazier.Item, error) {
	var i internal.Item

	tx, err := b.node.Begin(true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transaction")
	}
	defer tx.Rollback()

	err = tx.One("Key", key, &i)
	if err != nil {
		if err != storm.ErrNotFound {
			return nil, errors.Wrap(err, "failed to fetch item")
		}

		i = internal.Item{
			Key:  key,
			Data: data,
		}
	} else {
		i.Data = data
	}

	err = tx.Save(&i)
	if err != nil {
		return nil, err
	}

	return &brazier.Item{
		Key:  i.Key,
		Data: i.Data,
	}, tx.Commit()
}

// Get an item by key.
func (b *Bucket) Get(key string) (*brazier.Item, error) {
	var i internal.Item

	err := b.node.One("Key", key, &i)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, brazier.ErrKeyNotFound
		}

		return nil, errors.Wrap(err, "failed to fetch item")
	}

	return &brazier.Item{
		Key:  i.Key,
		Data: i.Data,
	}, nil
}

// Delete item from the bucket
func (b *Bucket) Delete(key string) error {
	var i internal.Item

	tx, err := b.node.Begin(true)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction")
	}
	defer tx.Rollback()

	err = tx.One("Key", key, &i)
	if err != nil {
		if err == storm.ErrNotFound {
			return brazier.ErrKeyNotFound
		}
		return errors.Wrap(err, "failed to fetch item")
	}

	err = tx.DeleteStruct(&i)
	if err != nil {
		return errors.Wrap(err, "failed to delete item")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit")
	}

	return nil
}

// Page returns a list of items
func (b *Bucket) Page(page int, perPage int) ([]brazier.Item, error) {
	var skip int
	var list []internal.Item

	if page <= 0 {
		return nil, nil
	}

	if perPage >= 0 {
		skip = (page - 1) * perPage
	}

	err := b.node.All(&list, storm.Skip(skip), storm.Limit(perPage))
	if err != nil {
		return nil, errors.Wrap(err, "boltdb.bucket.Page failed to fetch items")
	}

	items := make([]brazier.Item, len(list))
	for i := range list {
		items[i].Key = list[i].Key
		items[i].Data = list[i].Data
	}
	return items, nil
}

// Close the bucket session
func (b *Bucket) Close() error {
	return nil
}
