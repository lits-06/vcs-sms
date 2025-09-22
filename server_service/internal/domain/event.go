package domain

import "context"

type EventPublisher interface {
	PublishServerCreate(ctx context.Context, server *Server) error
	PublishServerDelete(ctx context.Context, serverID string) error
}