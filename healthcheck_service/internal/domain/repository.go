package domain

import (
	"context"
)

type Repository interface {
	GetServerSnapshot(serverID string) (*Server, error)

	GetAllServersSnapshot(ctx context.Context) (*[]Server, error)
}

type CacheRepository interface {
	GetServerState(ctx context.Context, serverID string) (*Server, error)
	SetServerState(ctx context.Context, serverID string, snapshot *Server) error
	DeleteServerState(ctx context.Context, serverID string) error

	GetAllServersSnapshot(ctx context.Context) (*[]Server, error)
	SetAllServersSnapshot(ctx context.Context, snapshots *[]Server) error
}
