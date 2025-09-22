package grpc

import (
	"google.golang.org/grpc"
)

type Config struct {
	Port string
}

func NewGRPCServer() *grpc.Server {
	return grpc.NewServer()
}
