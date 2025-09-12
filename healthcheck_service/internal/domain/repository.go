package domain

import (
	"context"
)

type Repository interface {
	GetServerSnapshot(serverID string) (*ServerSnapshot, error)
}

type StateRepository interface {
	GetServerSnapshot(ctx context.Context, serverID string) (*ServerSnapshot, error)
	SaveServerSnapshot(ctx context.Context, snapshot *ServerSnapshot) error
	CreateStateRecord(ctx context.Context, record *ServerState) error
}

type CacheRepository interface {
	GetServerState(ctx context.Context, serverID string) (*ServerSnapshot, error)
	SetServerState(ctx context.Context, serverID string, snapshot *ServerSnapshot) error
	DeleteServerState(ctx context.Context, serverID string) error
}
