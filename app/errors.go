package app

import (
	"bytes"
	"fmt"
)

type Errors []error

func (e Errors) Error() string {
	var buf bytes.Buffer

	for _, err := range e {
		fmt.Fprintf(&buf, "Err: %s\n", err.Error())
	}

	return buf.String()
}
