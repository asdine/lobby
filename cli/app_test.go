package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppConfigDir(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "lobby")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	f, err := ioutil.TempFile(dir, "lobby-")
	require.NoError(t, err)

	testCases := []map[string]interface{}{
		{"dir": dir, "error": false},                      // exists
		{"dir": path.Join(dir, "config"), "error": false}, // doesn't exist
		{"dir": f.Name(), "error": true},                  // file
	}

	err = f.Close()
	require.NoError(t, err)

	for _, test := range testCases {
		var out bytes.Buffer
		a := app{
			out:       &out,
			configDir: test["dir"].(string),
		}

		err = a.initConfigDir()
		require.Equal(t, test["error"].(bool), err != nil)
	}
}
