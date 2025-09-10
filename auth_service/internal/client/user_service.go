package client

import (
	"context"

	"github.com/lits-06/vcs-sms/auth_service/config"
	"google.golang.org/grpc"
)

func NewUserServiceConn(ctx context.Context, cfg *config.Config) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(cfg.Grpc.UserServicePort)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
