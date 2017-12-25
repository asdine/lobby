package rpc

import (
	"strings"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/validation"
	"google.golang.org/grpc"
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
func newError(err error, logger *log.Logger) error {
	var code codes.Code

	switch {
	case validation.IsError(err):
		code = codes.InvalidArgument
	case err == lobby.ErrTopicNotFound || err == lobby.ErrBackendNotFound:
		code = codes.NotFound
	case err == lobby.ErrTopicAlreadyExists:
		code = codes.AlreadyExists
	default:
		code = codes.Unknown
	}

	// Log error.
	logger.Debugf("grpc error: %s (code=%s)", err, code.String())

	// Hide error from client if it is internal.
	if code == codes.Unknown {
		err = ErrInternal
	}

	return status.Error(code, err.Error())
}

func errFromGRPC(err error) error {
	code := grpc.Code(err)

	switch code {
	case codes.AlreadyExists:
		return lobby.ErrTopicAlreadyExists
	case codes.NotFound:
		if strings.Contains(err.Error(), lobby.ErrBackendNotFound.Error()) {
			return lobby.ErrBackendNotFound
		}
		return lobby.ErrTopicNotFound
	default:
		return err
	}
}
