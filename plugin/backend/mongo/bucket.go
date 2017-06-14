package mongo

import (
	"encoding/json"

	"github.com/asdine/lobby"
	ljson "github.com/asdine/lobby/json"
	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const colItems = "items"

type item struct {
	ID     string      `bson:"_id,omitempty"`
	Bucket string      `bson:"bucket"`
	Key    string      `bson:"key"`
	Value  interface{} `bson:"value"`
}

var _ lobby.Bucket = new(Bucket)

// NewBucket returns a MongoDB Bucket.
func NewBucket(session *mgo.Session, name string) *Bucket {
	return &Bucket{
		session: session,
		name:    name,
	}
}

// Bucket is a MongoDB implementation of a bucket.
type Bucket struct {
	session *mgo.Session
	name    string
}

// Put value to the bucket. Returns an Item.
func (b *Bucket) Put(key string, value []byte) (*lobby.Item, error) {
	col := b.session.DB("").C(colItems)

	var raw interface{}

	valid, err := ljson.ValidateBytes(value)
	if err == nil {
		err := json.Unmarshal(valid, &raw)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal json")
		}
	} else {
		raw = value
	}

	_, err = col.Upsert(
		bson.M{"bucket": b.name, "key": key},
		bson.M{"$set": &item{Key: key, Bucket: b.name, Value: raw}})
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
	err := col.Find(bson.M{"bucket": b.name, "key": key}).One(&i)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, lobby.ErrKeyNotFound
		}

		return nil, errors.Wrap(err, "failed to fetch item")
	}

	value, ok := i.Value.([]byte)
	if !ok {
		value, err = json.Marshal(i.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal value into json")
		}
	}

	return &lobby.Item{
		Key:   i.Key,
		Value: value,
	}, nil
}

// Delete item from the bucket
func (b *Bucket) Delete(key string) error {
	col := b.session.DB("").C(colItems)
	err := col.Remove(bson.M{"bucket": b.name, "key": key})
	if err != nil {
		if err == mgo.ErrNotFound {
			return lobby.ErrKeyNotFound
		}

		return errors.Wrap(err, "failed to delete item")
	}

	return nil
}

// Close the bucket session
func (b *Bucket) Close() error {
	b.session.Close()
	return nil
}
