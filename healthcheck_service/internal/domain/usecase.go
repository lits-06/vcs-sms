package domain

import (
	"context"
)

type UseCase interface {
	StartHealthCheckScheduler(ctx context.Context)
	StopHealthCheckScheduler(ctx context.Context)
	CheckServersHealth(ctx context.Context, servers []Server)
	IndexServerState(ctx context.Context, server *Server) error
	DeleteServerSnapshot(ctx context.Context, serverID string) error

	BulkIndexServerStates(ctx context.Context, servers []Server) error
}
