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

type CreateServerRequest struct {
	ID     string              `json:"id" validate:"required"`
	Name   string              `json:"name" validate:"required"`
	IPv4   string              `json:"ipv4" validate:"required,ipv4"`
	Status entity.ServerStatus `json:"status" validate:"omitempty,oneof=ON OFF"`
}

type QueryServerRequest struct {
	Filter     ServerFilter     `json:"filter"`
	Pagination ServerPagination `json:"pagination"`
	Sort       ServerSort       `json:"sort"`
}

type QueryServerResponse struct {
	Servers *[]entity.Server `json:"servers"`
	Total   int              `json:"total"`
}

type UpdateServerRequest struct {
	ID     string              `json:"id"`
	Name   string              `json:"name,omitempty" validate:"omitempty"`
	IPv4   string              `json:"ipv4,omitempty" validate:"omitempty,ipv4"`
	Status entity.ServerStatus `json:"status,omitempty" validate:"omitempty,oneof=ON OFF"`
}

type ImportResponse struct {
	SuccessCount   int
	FailureCount   int
	SuccessServers []string // format: "ID:Name"
	FailureServers []string // format: "ID:Name - error message"
}
