package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/validation"
)

// HTTP errors
const (
	ErrInvalidJSON  = lobby.Error("invalid_json")
	ErrInternal     = lobby.Error("internal_error")
	ErrEmptyContent = lobby.Error("empty_content")
)

// Error writes an API error message to the response and logger.
func Error(w http.ResponseWriter, err error, code int, logger *log.Logger) {
	// Log error.
	logger.Printf("http error: %s (code=%d)", err, code)

	// Hide error from client if it is internal.
	if code == http.StatusInternalServerError {
		err = ErrInternal
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
