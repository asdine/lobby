package rpc

import (
	"log"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTP errors
const (
	ErrInvalidJSON  = lobby.Error("invalid_json")
	ErrInternal     = lobby.Error("internal_error")
	ErrEmptyContent = lobby.Error("empty_content")
)

// Error writes an API error message to the response and logger.
func Error(err error, logger *log.Logger) error {
	var code codes.Code

	switch {
	case validation.IsError(err):
		code = codes.InvalidArgument
	case err == lobby.ErrBucketNotFound || err == lobby.ErrKeyNotFound:
		code = codes.NotFound
	default:
		code = codes.Internal
	}

	// Log error.
	logger.Printf("grpc error: %s (code=%s)", err, code.String())

	// Hide error from client if it is internal.
	if code == codes.Internal {
		err = ErrInternal
	}

	return status.Error(code, err.Error())
}
