package domain

import (
	"context"

	"github.com/lits-06/vcs-sms/domain/entity"
)

// Repository defines the interface for server data operations
type Repository interface {
	// CRUD operations
	Create(ctx context.Context, server *Server) error
	GetByID(ctx context.Context, id string) (*Server, error)
	GetByName(ctx context.Context, name string) (*Server, error)
	Update(ctx context.Context, server *Server) error
	Delete(ctx context.Context, id string) error

	// Query operations
	List(ctx context.Context, filter ServerFilter, sort ServerSort, pagination ServerPagination) (*[]Server, int, error)

	// Validation operations
	ExistsWithID(ctx context.Context, id string) (bool, error)
	ExistsWithName(ctx context.Context, name string) (bool, error)
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
	Sort  string    `json:"sort,omitempty" validate:"omitempty,oneof=name status created_at updated_at" form:"sort"` // name, status, created_at, updated_at
	Order SortOrder `json:"order,omitempty" validate:"omitempty,oneof=asc desc" form:"order"`                        // asc, desc
}

// Pagination represents pagination parameters
type ServerPagination struct {
	From int `json:"from,omitempty" form:"from"` // offset
	To   int `json:"to,omitempty" form:"to"`     // limit (or you can use Size instead)
}
