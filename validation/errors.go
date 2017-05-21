package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type validationError map[string][]error

func (e validationError) Error() string {
	var buf bytes.Buffer

	i := 0
	for k, errs := range e {
		if i > 0 {
			fmt.Fprintf(&buf, "; ")
		}

		for j := range errs {
			if j == 0 {
				fmt.Fprintf(&buf, "%s: ", k)
			} else if j > 0 {
				fmt.Fprintf(&buf, ",")
			}

			fmt.Fprintf(&buf, "%s", errs[j])
		}

		i++
	}

	return buf.String()
}

// MarshalJSON creates a JSON representation of the validation error.
func (e validationError) MarshalJSON() ([]byte, error) {
	s := make(map[string][]string)

	for k, errs := range e {
		s[k] = make([]string, len(errs))
		for i := range errs {
			s[k][i] = errs[i].Error()
		}
	}

	return json.Marshal(s)
}

// AddError adds an error under the given name.
func AddError(verr error, name string, err error) error {
	var e validationError
	var ok bool

	if verr != nil {
		e, ok = verr.(validationError)
		if !ok {
			panic("incompatible error")
		}
	}

	if e == nil {
		e = make(validationError)
	}

	e[name] = append(e[name], err)
	return e
}

// LastError returns the last error for the specified field.
func LastError(err error, name string) error {
	e, ok := err.(validationError)
	if !ok {
		return nil
	}

	errs, ok := e[name]
	if !ok || len(errs) == 0 {
		return nil
	}
	return errs[len(errs)-1]
}

// IsError tells if the given error is a validation error.
func IsError(err error) bool {
	_, ok := err.(validationError)
	return ok
}
