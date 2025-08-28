package domain

import (
	"context"
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
