package domain

import (
	"context"
	"mime/multipart"
)

type UseCase interface {
	CreateServer(ctx context.Context, req CreateServerRequest) (*Server, error)
	ViewServer(ctx context.Context, req QueryServerRequest) (*QueryServerResponse, error)
	UpdateServer(ctx context.Context, req UpdateServerRequest) error
	DeleteServer(ctx context.Context, serverID string) error

	ImportServersFromExcel(ctx context.Context, file multipart.File) (*ImportResponse, error)
	ExportServersToExcel(ctx context.Context, req QueryServerRequest) error
}
