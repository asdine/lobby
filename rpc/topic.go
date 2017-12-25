package rpc

import (
	"context"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/asdine/lobby/validation"
)

func newTopicService(b lobby.Backend, logger *log.Logger) *topicService {
	return &topicService{
		backend: b,
		logger:  logger,
	}
}

type topicService struct {
	backend lobby.Backend
	logger  *log.Logger
}

// Send an message to a topic.
func (s *topicService) Send(ctx context.Context, message *proto.NewMessage) (*proto.Empty, error) {
	err := validation.Validate(message)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	t, err := s.backend.Topic(message.Topic)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	err = t.Send(&lobby.Message{
		Group: message.Message.Group,
		Value: message.Message.Value,
	})
	if err != nil {
		return nil, newError(err, s.logger)
	}

	return new(proto.Empty), nil
}
