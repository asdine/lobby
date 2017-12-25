package http

import (
	"encoding/json"
	"net/http"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/validation"
)

// HTTP errors
const (
	errInvalidJSON  = lobby.Error("invalid_json")
	errInternal     = lobby.Error("internal_error")
	errEmptyContent = lobby.Error("empty_content")
)

// writeError writes an API error message to the response and logger.
func writeError(w http.ResponseWriter, err error, code int, logger *log.Logger) {
	// Log error.
	logger.Debugf("http error: %s (code=%d)", err, code)

	// Hide error from client if it is internal.
	if code == http.StatusInternalServerError {
		err = errInternal
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	enc := json.NewEncoder(w)
	switch {
	case validation.IsError(err):
		err = enc.Encode(&validationErrorResponse{
			Err:    "validation error",
			Fields: err,
		})
	default:
		err = enc.Encode(&errorResponse{Err: err.Error()})
	}

	if err != nil {
		logger.Println(err)
	}
}

// errorResponse is a generic response for sending an error.
type errorResponse struct {
	Err string `json:"err,omitempty"`
}

// validationErrorResponse is used for validation errors.
type validationErrorResponse struct {
	Err    string `json:"err,omitempty"`
	Fields error  `json:"fields"`
}
