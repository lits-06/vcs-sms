package domain

import (
	"context"
)

type Repository interface {
	SaveServerSnapshot(ctx context.Context, server *Server) error
	IndexServerState(ctx context.Context, server *Server) error
	GetAllServersSnapshot(ctx context.Context) ([]Server, error)
	DeleteServerSnapshot(ctx context.Context, serverID string) error

	BulkIndexServerStates(ctx context.Context, servers []Server) error
	BulkSaveServerSnapshots(ctx context.Context, servers []Server) error
}

type CacheRepository interface {
	GetServerSnapshot(ctx context.Context, serverID string) (*Server, error)
	SaveServerSnapshot(ctx context.Context, server *Server) error
	GetAllServersSnapshot(ctx context.Context) ([]Server, error)
	SetAllServersSnapshot(ctx context.Context, servers []Server) error
	DeleteServerSnapshot(ctx context.Context, serverID string) error
}
