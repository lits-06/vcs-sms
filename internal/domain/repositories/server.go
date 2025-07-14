package repositories

import (
	"context"

	"github.com/lits-06/vcs-sms/internal/domain/entities"
)

// ServerRepository defines the interface for server data operations
type ServerRepository interface {
	// CRUD operations
	Create(ctx context.Context, server *entities.Server) error
	GetByID(ctx context.Context, id string) (*entities.Server, error)
	GetByName(ctx context.Context, name string) (*entities.Server, error)
	Update(ctx context.Context, server *entities.Server) error
	Delete(ctx context.Context, id string) error

	// Query operations
	List(ctx context.Context, filter entities.ServerFilter, sort entities.ServerSort, pagination entities.ServerPagination) ([]entities.Server, int, error)

	// Validation operations
	ExistsWithID(ctx context.Context, id string) (bool, error)
	ExistsWithName(ctx context.Context, name string) (bool, error)
}
