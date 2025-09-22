package domain

import (
	"context"
)

// Repository defines the interface for server data operations
type Repository interface {
	// CRUD operations
	Create(ctx context.Context, server *Server) (*Server, error)
	GetByID(ctx context.Context, id string) (*Server, error)
	GetByName(ctx context.Context, name string) (*Server, error)
	Update(ctx context.Context, id, name, ipv4 string) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status string) error

	// Query operations
	List(ctx context.Context, name, status, ipv4 string, from, to int, sort, order string) (*[]Server, int, error)

	// Validation operations
	ExistsWithID(ctx context.Context, id string) (bool, error)
	ExistsWithName(ctx context.Context, name string) (bool, error)
}
