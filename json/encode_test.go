package json_test

import (
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/json"
	"github.com/stretchr/testify/require"
)

func TestMarshalList(t *testing.T) {
	items := []lobby.Item{
		{Key: "k1", Data: []byte(`"Data1"`)},
		{Key: "k2", Data: []byte(`"Data2"`)},
	}

	expected := `[{"key":"k1","value":"Data1"},{"key":"k2","value":"Data2"}]`
	out, err := json.MarshalList(items)
	require.NoError(t, err)
	require.Equal(t, expected, string(out))
}
