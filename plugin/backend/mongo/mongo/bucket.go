package mongo

import (
	"github.com/asdine/lobby"
	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const colItems = "items"

type item struct {
	ID    string   `bson:"_id"`
	Key   string   `bson:"key"`
	Value bson.Raw `bson:"value"`
}

var _ lobby.Bucket = new(Bucket)

// NewBucket returns a MongoDB Bucket.
func NewBucket(session *mgo.Session) *Bucket {
	return &Bucket{
		session: session,
	}
}

// Bucket is a MongoDB implementation of a bucket.
type Bucket struct {
	session *mgo.Session
}

// Put value to the bucket. Returns an Item.
func (b *Bucket) Put(key string, value []byte) (*lobby.Item, error) {
	col := b.session.DB("").C(colItems)

	_, err := col.Upsert(bson.M{"key": key}, &item{Key: key, Value: bson.Raw{Data: value}})
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert of update")
	}

	return &lobby.Item{
		Key:   key,
		Value: value,
	}, nil
}

// Get an item by key.
func (b *Bucket) Get(key string) (*lobby.Item, error) {
	var i item

	col := b.session.DB("").C(colItems)
	err := col.Find(bson.M{"key": key}).One(&i)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, lobby.ErrKeyNotFound
		}

		return nil, errors.Wrap(err, "failed to fetch item")
	}

	return &lobby.Item{
		Key:   i.Key,
		Value: i.Value.Data,
	}, nil
}

// Delete item from the bucket
func (b *Bucket) Delete(key string) error {
	col := b.session.DB("").C(colItems)
	err := col.Remove(bson.M{"key": key})
	if err != nil {
		if err == mgo.ErrNotFound {
			return lobby.ErrKeyNotFound
		}

		return errors.Wrap(err, "failed to delete item")
	}

	return nil
}

// Page returns a list of items
func (b *Bucket) Page(page int, perPage int) ([]lobby.Item, error) {
	var skip int
	var list []item

	if page <= 0 {
		return nil, nil
	}

	if perPage >= 0 {
		skip = (page - 1) * perPage
	}

	col := b.session.DB("").C(colItems)
	err := col.Find(nil).Skip(skip).Limit(perPage).Sort("_id").All(&list)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch items")
	}

	items := make([]lobby.Item, len(list))
	for i := range list {
		items[i].Key = list[i].Key
		items[i].Value = list[i].Value.Data
	}
	return items, nil
}

// Close the bucket session
func (b *Bucket) Close() error {
	b.session.Close()
	return nil
}
