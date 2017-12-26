package app

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaths(t *testing.T) {
	t.Run("Empty Config dir", func(t *testing.T) {
		var p Paths
		err := p.Create()
		require.Error(t, err)
	})

	t.Run("Bad dir", func(t *testing.T) {
		p := Paths{
			DataDir:   "/some path",
			SocketDir: "/some path",
		}
		err := p.Create()
		require.Error(t, err)
	})

	t.Run("Exist As File", func(t *testing.T) {
		name, err := ioutil.TempDir("", "lobby")
		require.NoError(t, err)
		defer os.RemoveAll(name)

		f, err := os.Create(path.Join(name, "config"))
		require.NoError(t, err)
		err = f.Close()
		require.NoError(t, err)

		p := Paths{
			DataDir:   path.Join(name, "config"),
			SocketDir: path.Join(name, "config", "sockets"),
		}
		err = p.Create()
		require.Error(t, err)
	})

	okTest := func(t *testing.T, fn func(string) *Paths) {
		name, err := ioutil.TempDir("", "lobby")
		require.NoError(t, err)
		defer os.RemoveAll(name)

		p := fn(name)

		err = p.Create()
		require.NoError(t, err)

		_, err = os.Stat(p.DataDir)
		require.NoError(t, err)

		_, err = os.Stat(p.SocketDir)
		require.NoError(t, err)

		err = p.Create()
		require.NoError(t, err)
	}

	t.Run("OK", func(t *testing.T) {
		okTest(t, func(name string) *Paths {
			return &Paths{
				DataDir:   path.Join(name, "config"),
				SocketDir: path.Join(name, "config", "sockets"),
			}
		})
	})

	t.Run("No socket dir", func(t *testing.T) {
		okTest(t, func(name string) *Paths {
			return &Paths{
				DataDir: path.Join(name, "config"),
			}
		})
	})
}
