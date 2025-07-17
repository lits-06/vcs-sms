package server

import (
	"context"

	"github.com/lits-06/vcs-sms/entity"
)

type ServerProvider interface {
	CreateServer(ctx context.Context, server *entity.Server) error
	DeleteServer(ctx context.Context, serverID string) error
	StartServer(ctx context.Context, serverID string) error
	StopServer(ctx context.Context, serverID string) error
	GetServerStatus(ctx context.Context, serverID string) (entity.ServerStatus, error)
	ListServers(ctx context.Context) (*[]entity.Server, error)
}
