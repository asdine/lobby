package rpc

import (
	"log"

	"github.com/asdine/lobby"
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

	switch err {
	case lobby.ErrBucketNotFound:
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

	return status.Error(codes.Internal, err.Error())
}
