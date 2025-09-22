package grpc

import (
	"google.golang.org/grpc"
)

type Config struct {
	Port string `mapstructure:"port"`
}

func NewGRPCServer() *grpc.Server {
	return grpc.NewServer()
}
