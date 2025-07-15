package server

import (
	"context"

	"github.com/lits-06/vcs-sms/entity"
)

type UseCase interface {
	CreateServer(ctx context.Context, req CreateServerRequest) (*entity.Server, error)
	ViewServer(ctx context.Context, req ViewServerRequest) (*ViewServerResponse, error)
	UpdateServer(ctx context.Context, req UpdateServerRequest) error
	DeleteServer(ctx context.Context, serverID string) error
}

// Repository defines the interface for server data operations
type Repository interface {
	// CRUD operations
	Create(ctx context.Context, server *entity.Server) error
	GetByID(ctx context.Context, id string) (*entity.Server, error)
	GetByName(ctx context.Context, name string) (*entity.Server, error)
	Update(ctx context.Context, server *entity.Server) error
	Delete(ctx context.Context, id string) error

	// Query operations
	List(ctx context.Context, filter ServerFilter, sort ServerSort, pagination ServerPagination) ([]entity.Server, int, error)

	// Validation operations
	ExistsWithID(ctx context.Context, id string) (bool, error)
	ExistsWithName(ctx context.Context, name string) (bool, error)
}

// ServerFilter represents filtering criteria for servers
type ServerFilter struct {
	Name   string              `json:"name,omitempty" validate:"omitempty"`
	Status entity.ServerStatus `json:"status,omitempty" validate:"omitempty,oneof=ON OFF UNKNOWN"`
	IPv4   string              `json:"ipv4,omitempty"	validate:"omitempty,ipv4"`
}

// SortOrder represents sorting direction
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// ServerSort represents sorting criteria
type ServerSort struct {
	Field string    `json:"field,omitempty" validate:"omitempty,oneof=name status created_at updated_at last_checked"`
	Order SortOrder `json:"order,omitempty" validate:"omitempty,oneof=asc desc"` // asc, desc
}

// Pagination represents pagination parameters
type ServerPagination struct {
	From  int `json:"from,omitempty"`  // offset
	To    int `json:"to,omitempty"`    // limit (or you can use Size instead)
	Size  int `json:"size,omitempty"`  // page size
	Page  int `json:"page,omitempty"`  // page number (alternative to from/to)
	Total int `json:"total,omitempty"` // total count (for response)
}

type CreateServerRequest struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
	IPv4 string `json:"ipv4" validate:"required,ipv4"`
}

type ViewServerRequest struct {
	Filter     ServerFilter     `json:"filter"`
	Pagination ServerPagination `json:"pagination"`
	Sort       ServerSort       `json:"sort"`
}

type ViewServerResponse struct {
	Servers []entity.Server `json:"servers"`
	Total   int             `json:"total"`
}

type UpdateServerRequest struct {
	ServerID   string                 `json:"id"`
	UpdateData map[string]interface{} `json:"update_data"`
}
