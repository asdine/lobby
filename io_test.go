package lobby_test

import (
	"bytes"
	"testing"

	"github.com/asdine/lobby"
	"github.com/stretchr/testify/require"
)

func TestPrefixWriter(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		var buf bytes.Buffer
		p := lobby.NewPrefixWriter("[prefix] ", &buf)
		n, err := p.Write([]byte(""))
		require.NoError(t, err)
		require.Zero(t, n)
		require.Zero(t, buf.Len())
	})

	t.Run("Simple line", func(t *testing.T) {
		var buf bytes.Buffer
		p := lobby.NewPrefixWriter("[prefix] ", &buf)
		src := []byte("Hello\n")
		n, err := p.Write(src)
		require.NoError(t, err)
		result := "[prefix] Hello\n"
		require.Equal(t, len(src), n)
		require.Equal(t, result, buf.String())
	})

	t.Run("Multi part line", func(t *testing.T) {
		var buf bytes.Buffer
		p := lobby.NewPrefixWriter("[prefix] ", &buf)
		src := []byte("Hello")
		n, err := p.Write(src)
		require.NoError(t, err)
		require.Equal(t, len(src), n)
		require.Zero(t, buf.Len())

		src = []byte(" World\nHow are")
		n, err = p.Write(src)
		require.NoError(t, err)
		result := "[prefix] Hello World\n"
		require.Equal(t, len(src), n)
		require.Equal(t, result, buf.String())
		buf.Reset()

		src = []byte(" you ?\n")
		n, err = p.Write(src)
		require.NoError(t, err)
		result = "[prefix] How are you ?\n"
		require.Equal(t, len(src), n)
		require.Equal(t, result, buf.String())
	})

	t.Run("Multiline at once", func(t *testing.T) {
		var buf bytes.Buffer
		p := lobby.NewPrefixWriter("[prefix] ", &buf)
		src := []byte("Hello World\nHow are you ?\nI'm")
		n, err := p.Write(src)
		require.NoError(t, err)
		result := "[prefix] Hello World\n[prefix] How are you ?\n"
		require.Equal(t, len(src), n)
		require.Equal(t, result, buf.String())
		buf.Reset()

		src = []byte(" fine\n")
		n, err = p.Write(src)
		require.NoError(t, err)
		result = "[prefix] I'm fine\n"
		require.Equal(t, len(src), n)
		require.Equal(t, result, buf.String())
	})

	t.Run("Skip empty lines", func(t *testing.T) {
		var buf bytes.Buffer
		p := lobby.NewPrefixWriter("[prefix] ", &buf)
		src := []byte("\n")
		n, err := p.Write(src)
		require.NoError(t, err)
		require.Equal(t, len(src), n)
		require.Zero(t, buf.Len())

		src = []byte("\n\n\n")
		n, err = p.Write(src)
		require.NoError(t, err)
		require.Equal(t, len(src), n)
		require.Zero(t, buf.Len())

		src = []byte("Hello\n\n\n")
		n, err = p.Write(src)
		require.NoError(t, err)
		result := "[prefix] Hello\n"
		require.Equal(t, len(src), n)
		require.Equal(t, result, buf.String())
		buf.Reset()

		src = []byte("Hello\n\n\n")
		n, err = p.Write(src)
		require.NoError(t, err)
		result = "[prefix] Hello\n"
		require.Equal(t, len(src), n)
		require.Equal(t, result, buf.String())
	})
}
