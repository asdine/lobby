package json

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode"
	"unicode/utf8"
)

// ValidateBytes checks if the data is valid json.
func ValidateBytes(data []byte) ([]byte, error) {
	var i json.RawMessage

	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// Clean removes unnecessary white space from a json value.
func Clean(data []byte) []byte {
	to := make([]byte, len(data))

	var inString bool
	start := 0
	i := 0
	l := len(data)
	for start < l {
		wid := 1
		r := rune(data[start])
		if r >= utf8.RuneSelf {
			r, wid = utf8.DecodeRune(data[start:])
		}
		start += wid

		if r == '"' {
			inString = !inString
		} else if unicode.IsSpace(r) {
			if !inString {
				continue
			}
		}
		i += utf8.EncodeRune(to[i:], r)
	}

	return to[:i]
}

// ToValidJSONFromReader converts the content of a reader to a valid json value.
func ToValidJSONFromReader(r io.Reader) []byte {
	var buffer bytes.Buffer
	n, err := buffer.ReadFrom(r)
	if n == 0 || err != nil {
		return nil
	}

	return ToValidJSONFromBytes(buffer.Bytes())
}

// ToValidJSONFromBytes converts data to a valid json value.
func ToValidJSONFromBytes(data []byte) []byte {
	d, err := ValidateBytes(data)
	if err == nil {
		return Clean(d)
	}

	return WrapIntoJSONString(data)
}

// WrapIntoJSONString turns any data into a valid json value.
func WrapIntoJSONString(data []byte) []byte {
	count := bytes.Count(data, []byte(`"`))
	out := make([]byte, len(data)+count+2)

	out[0] = '"'
	j := 1
	for _, b := range data {
		if b == '"' {
			out[j] = '\\'
			j++
		}
		out[j] = b
		j++
	}
	out[j] = '"'
	return out
}
