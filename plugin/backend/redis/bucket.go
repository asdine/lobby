package redis

import (
	"github.com/asdine/lobby"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

var _ lobby.Bucket = new(Bucket)

// NewBucket returns a Redis Bucket.
func NewBucket(conn redis.Conn, name string) *Bucket {
	return &Bucket{
		conn: conn,
		name: name,
	}
}

// Bucket is a Redis implementation of a bucket.
type Bucket struct {
	conn redis.Conn
	name string
}

// Put value to the bucket. Returns an Item.
func (b *Bucket) Put(key string, value []byte) (*lobby.Item, error) {
	_, err := b.conn.Do("HSET", b.name, key, value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to put item")
	}

	return &lobby.Item{
		Key:   key,
		Value: value,
	}, nil
}

// Get an item by key.
func (b *Bucket) Get(key string) (*lobby.Item, error) {
	value, err := redis.Bytes(b.conn.Do("HGET", b.name, key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, lobby.ErrKeyNotFound
		}

		return nil, errors.Wrap(err, "failed to fetch item")
	}

	return &lobby.Item{
		Key:   key,
		Value: value,
	}, nil
}

// Delete item from the bucket.
func (b *Bucket) Delete(key string) error {
	ret, err := redis.Int(b.conn.Do("HDEL", b.name, key))
	if err != nil {
		return errors.Wrap(err, "failed to delete item")
	}

	if ret == 0 {
		return lobby.ErrKeyNotFound
	}

	return nil
}

// Close the bucket connection
func (b *Bucket) Close() error {
	return b.conn.Close()
}
