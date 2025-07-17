package server

import (
	"context"
	"mime/multipart"

	"github.com/lits-06/vcs-sms/entity"
)

type UseCase interface {
	CreateServer(ctx context.Context, req CreateServerRequest) (*entity.Server, error)
	ViewServer(ctx context.Context, req QueryServerRequest) (*QueryServerResponse, error)
	UpdateServer(ctx context.Context, req UpdateServerRequest) error
	DeleteServer(ctx context.Context, serverID string) error

	ImportServersFromExcel(ctx context.Context, file multipart.File) (*ImportRespose, error)
	ExportServersToExcel(ctx context.Context, req QueryServerRequest) error
}

// ServerFilter represents filtering criteria for servers
type ServerFilter struct {
	Name   string              `json:"name,omitempty" validate:"omitempty" form:"name"`
	Status entity.ServerStatus `json:"status,omitempty" validate:"omitempty,oneof=ON OFF" form:"status"`
	IPv4   string              `json:"ipv4,omitempty" validate:"omitempty,ipv4" form:"ipv4"`
}

// SortOrder represents sorting direction
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// ServerSort represents sorting criteria
type ServerSort struct {
	Sort  string    `json:"sort,omitempty" validate:"omitempty,oneof=name status created_at updated_at last_checked" form:"sort"` // name, status, created_at, updated_at, last_checked
	Order SortOrder `json:"order,omitempty" validate:"omitempty,oneof=asc desc" form:"order"`                                     // asc, desc
}

// Pagination represents pagination parameters
type ServerPagination struct {
	From int `json:"from,omitempty" form:"from"` // offset
	To   int `json:"to,omitempty" form:"to"`     // limit (or you can use Size instead)
}

type CreateServerRequest struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
	IPv4 string `json:"ipv4" validate:"required,ipv4"`
}

type QueryServerRequest struct {
	Filter     ServerFilter     `json:"filter"`
	Pagination ServerPagination `json:"pagination"`
	Sort       ServerSort       `json:"sort"`
}

type QueryServerResponse struct {
	Servers []entity.Server `json:"servers"`
	Total   int             `json:"total"`
}

type UpdateServerRequest struct {
	ID     string              `json:"id"`
	Name   string              `json:"name,omitempty" validate:"omitempty"`
	Status entity.ServerStatus `json:"status,omitempty" validate:"omitempty,oneof=ON OFF"`
	IPv4   string              `json:"ipv4,omitempty" validate:"omitempty,ipv4"`
}

type ImportRespose struct {
	SuccessCount   int
	FailureCount   int
	SuccessServers []string // format: "ID:Name"
	FailureServers []string // format: "ID:Name - error message"
}
