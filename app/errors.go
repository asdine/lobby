package app

import (
	"bytes"
	"fmt"
)

// Errors contains a list of errors stored during the lifecycle of the App.
type Errors []error

func (e Errors) Error() string {
	var buf bytes.Buffer

	for i, err := range e {
		if i > 0 {
			fmt.Fprintf(&buf, "\n")
		}
		fmt.Fprintf(&buf, "Err: %s", err.Error())
	}

	return buf.String()
}
