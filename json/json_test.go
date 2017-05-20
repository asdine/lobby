package json_test

import (
	"strings"
	"testing"

	"github.com/asdine/lobby/json"
	"github.com/stretchr/testify/require"
)

func TestValidateFromBytes(t *testing.T) {
	data, err := json.ValidateBytes([]byte(""))
	require.Error(t, err)
	require.Nil(t, data)
	data, err = json.ValidateBytes([]byte("   "))
	require.Error(t, err)
	require.Nil(t, data)
	data, err = json.ValidateBytes([]byte(`"string"`))
	require.NoError(t, err)
	require.Equal(t, `"string"`, string(data))
	data, err = json.ValidateBytes([]byte(`10.6`))
	require.NoError(t, err)
	require.Equal(t, `10.6`, string(data))
	data, err = json.ValidateBytes([]byte(`{"key": "value"}`))
	require.NoError(t, err)
	require.Equal(t, `{"key": "value"}`, string(data))
	data, err = json.ValidateBytes([]byte(`{"bad": "format"`))
	require.Error(t, err)
	require.Nil(t, data)
	data, err = json.ValidateBytes([]byte(`something else`))
	require.Error(t, err)
	require.Nil(t, data)
}

func TestClean(t *testing.T) {
	require.Equal(t, []byte(``), json.Clean([]byte(``)))
	require.Equal(t, []byte(``), json.Clean([]byte(`      
	   `)))
	require.Equal(t, []byte(`"a b    c"`), json.Clean([]byte(`"a b    c"`)))
	require.Equal(t, []byte(`"a b    c"`), json.Clean([]byte(`   "a b    c"  `)))
	require.Equal(t, []byte(`10`), json.Clean([]byte("10\n")))
	require.Equal(t, []byte(`{"the name":"  &éà","another     key":[1,10,9,"    str  "]}`), json.Clean([]byte(`

		{
								"the name"       : "  &éà"      , "another     key"   : [ 1,  		10,9, "    str  " ]   }


		`)))
}

func TestToValidJSON(t *testing.T) {
	tests := map[string]string{
		`invalid è`:                    `"invalid è"`,
		`{"invalid": "json"`:           `"{\"invalid\": \"json\""`,
		`"valid"`:                      `"valid"`,
		`{"dirty"      :   "json" }  `: `{"dirty":"json"}`,
		`5`: `5`,
	}

	for in, out := range tests {
		res := json.ToValidJSONFromBytes([]byte(in))
		require.Equal(t, out, string(res))
	}

	for in, out := range tests {
		res := json.ToValidJSONFromReader(strings.NewReader(in))
		require.Equal(t, out, string(res))
	}
}

func BenchmarkClean(b *testing.B) {
	data := []byte(`

		{
								"the name"       : "  &éà"      , "another     key"   : [ 1,  		10,9, "    str  " ]   }


		`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Clean(data)
	}
}

func BenchmarkToValidJSON(b *testing.B) {
	invalidJSON := []byte(`

		{
								"the name"       : "  &éà"      , "another     key"   : [ 1,  		10,9, "    str  " ]


		`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.ToValidJSONFromBytes(invalidJSON)
	}
}
