package domain

import (
	"context"
	"mime/multipart"
)

type UseCase interface {
	CreateServer(ctx context.Context, name, ipv4 string, port int) (*Server, error)
	ViewServer(ctx context.Context, name, status, ipv4 string, from, to int, sort, order string) (*[]Server, int, error)
	UpdateServer(ctx context.Context, id, name, ipv4 string) error
	DeleteServer(ctx context.Context, serverID string) error
	UpdateServerStatus(ctx context.Context, serverID string, status string) error
	BulkUpdateServerStatus(ctx context.Context, updates []ServerStatusUpdate) error
	ImportServersFromExcel(ctx context.Context, file multipart.File) (*ImportResponse, error)
	ExportServersToExcel(ctx context.Context, name, status, ipv4 string, from, to int, sort, order string) ([]byte, error)
}
