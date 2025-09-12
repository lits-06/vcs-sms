package domain

import "context"

type EventPublisher interface {
	PublishStateChange(ctx context.Context, state *ServerState) error
}
