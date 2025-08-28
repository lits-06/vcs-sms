package domain

import (
	"context"
	"mime/multipart"

	"github.com/lits-06/vcs-sms/domain/entity"
)

type UseCase interface {
	CreateServer(ctx context.Context, req CreateServerRequest) (*Server, error)
	ViewServer(ctx context.Context, req QueryServerRequest) (*QueryServerResponse, error)
	UpdateServer(ctx context.Context, req UpdateServerRequest) error
	DeleteServer(ctx context.Context, serverID string) error

	ImportServersFromExcel(ctx context.Context, file multipart.File) (*ImportResponse, error)
	ExportServersToExcel(ctx context.Context, req QueryServerRequest) error
}

type QueryServerResponse struct {
	Servers *[]entity.Server `json:"servers"`
	Total   int              `json:"total"`
}

type ImportResponse struct {
	SuccessCount   int
	FailureCount   int
	SuccessServers []string // format: "ID:Name"
	FailureServers []string // format: "ID:Name - error message"
}
