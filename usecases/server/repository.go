package server

import (
	"context"

	"github.com/lits-06/vcs-sms/entity"
)

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
