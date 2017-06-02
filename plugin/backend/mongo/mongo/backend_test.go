package mongo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type errorHandler interface {
	Error(args ...interface{})
}

func newBackend(t errorHandler) (*Backend, func()) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017/test-db"
	}

	bck, err := NewBackend(uri)
	if err != nil {
		t.Error(err)
	}

	return bck, func() {
		_, err := bck.session.DB("").C(colItems).RemoveAll(nil)
		if err != nil {
			t.Error(err)
		}

		bck.Close()
	}
}

func TestBackend(t *testing.T) {
	backend, cleanup := newBackend(t)
	defer cleanup()

	bucket, err := backend.Bucket("a")
	require.NoError(t, err)
	require.NotNil(t, bucket)
	require.NotNil(t, bucket.(*Bucket).session)

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
