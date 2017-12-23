package bolt_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/lobby/bolt"
	"github.com/stretchr/testify/require"
)

type errorHandler interface {
	Error(args ...interface{})
}

func preparePath(t errorHandler, dbName string) (string, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "lobby")
	if err != nil {
		t.Error(err)
	}

	return filepath.Join(dir, dbName), func() {
		os.RemoveAll(dir)
	}
}

func TestBackend(t *testing.T) {
	path, cleanup := preparePath(t, "backend.db")
	defer cleanup()

	s, err := bolt.NewBackend(path)
	require.NoError(t, err)
	defer s.Close()

	topic, err := s.Topic("a")
	require.NoError(t, err)
	require.NotNil(t, topic)
	require.NotNil(t, s.DB)

	err = topic.Close()
	require.NoError(t, err)

	b1, err := s.Topic("a")
	require.NoError(t, err)

	b2, err := s.Topic("b")
	require.NoError(t, err)

	err = b1.Close()
	require.NoError(t, err)

	err = b2.Close()
	require.NoError(t, err)
}
