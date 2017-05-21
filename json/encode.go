package json

import (
	"encoding/json"

	"github.com/asdine/lobby"
)

// MarshalList marshals a list of items.
func MarshalList(items []lobby.Item) ([]byte, error) {
	return json.Marshal(marshalList(items))
}

// MarshalListPretty marshals a list of items.
func MarshalListPretty(items []lobby.Item) ([]byte, error) {
	return json.MarshalIndent(marshalList(items), "", "  ")
}

func marshalList(items []lobby.Item) []map[string]interface{} {
	list := make([]map[string]interface{}, len(items))

	for i := range items {
		k := items[i].Key
		list[i] = map[string]interface{}{
			"key": k,
		}

		v := json.RawMessage(items[i].Value)
		list[i]["value"] = &v
	}

	return list
}
