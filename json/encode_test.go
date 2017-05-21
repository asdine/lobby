package json_test

import (
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/json"
	"github.com/stretchr/testify/require"
)

func TestMarshalList(t *testing.T) {
	items := []lobby.Item{
		{Key: "k1", Value: []byte(`"Value1"`)},
		{Key: "k2", Value: []byte(`"Value2"`)},
	}

	expected := `[{"key":"k1","value":"Value1"},{"key":"k2","value":"Value2"}]`
	out, err := json.MarshalList(items)
	require.NoError(t, err)
	require.Equal(t, expected, string(out))
}
