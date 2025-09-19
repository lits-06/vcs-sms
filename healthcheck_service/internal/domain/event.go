package domain

import "context"

type EventPublisher interface {
	PublishStateChange(ctx context.Context, server *Server) error
}
