package redis

import (
	"time"

	"github.com/asdine/lobby"
	"github.com/garyburd/redigo/redis"
)

var _ lobby.Backend = new(Backend)

// NewBackend returns a Redis backend.
func NewBackend(addr string) (*Backend, error) {
	pool := redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	conn := pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return nil, err
	}

	return &Backend{
		pool: &pool,
	}, nil
}

// Backend is a Redis backend.
type Backend struct {
	pool *redis.Pool
}

// Topic returns the topic associated with the given name.
func (s *Backend) Topic(name string) (lobby.Topic, error) {
	return NewTopic(s.pool.Get(), name), nil
}

// Close the Redis connection.
func (s *Backend) Close() error {
	return s.pool.Close()
}
