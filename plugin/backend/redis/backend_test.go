package redis

import (
	"fmt"
	"log"
	"os"
	"testing"

	dockertest "gopkg.in/ory-am/dockertest.v3"

	"github.com/stretchr/testify/require"
)

var bck *Backend

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "3.2.9",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error

		bck, err = NewBackend(fmt.Sprintf(":%s", resource.GetPort("6379/tcp")))
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

type errorHandler interface {
	Error(args ...interface{})
}

func getBackend(t errorHandler) (*Backend, func()) {
	return bck, func() {
		conn := bck.pool.Get()
		defer conn.Close()
		_, err := conn.Do("FLUSHDB")
		if err != nil {
			t.Error(err)
		}
	}
}

func TestBackend(t *testing.T) {
	backend, cleanup := getBackend(t)
	defer cleanup()

	bucket, err := backend.Bucket("a")
	require.NoError(t, err)
	require.NotNil(t, bucket)
	require.NotNil(t, bucket.(*Bucket).conn)

	err = bucket.Close()
	require.NoError(t, err)

	b1, err := backend.Bucket("a")
	require.NoError(t, err)

	b2, err := backend.Bucket("b")
	require.NoError(t, err)

	err = b1.Close()
	require.NoError(t, err)

	err = b2.Close()
	require.NoError(t, err)
}
