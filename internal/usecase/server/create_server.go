package server

import (
	"context"
	"fmt"
	"time"

	"github.com/lits-06/vcs-sms/internal/domain/entities"
	"github.com/lits-06/vcs-sms/internal/domain/repositories"
)

type CreateServerUseCase struct {
	serverRepo repositories.ServerRepository
}

func NewCreateServerUseCase(serverRepo repositories.ServerRepository) *CreateServerUseCase {
	return &CreateServerUseCase{
		serverRepo: serverRepo,
	}
}

type CreateServerRequest struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
	IPv4 string `json:"ipv4" validate:"required,ipv4"`
}

func (uc *CreateServerUseCase) CreateServer(ctx context.Context, req CreateServerRequest) (*entities.Server, error) {
	// Validate request
	// if req.Name == "" || req.IPv4 == "" {
	// 	return nil, fmt.Errorf("name and IPv4 are required")
	// }

	// ipv4 invalid

	// Check if server with same ID already exists
	exist, err := uc.serverRepo.ExistsWithID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if server ID exists: %w", err)
	}
	if exist {
		return nil, fmt.Errorf("server with ID %s already exists", req.ID)
	}

	// Check if server with same name already exists
	exist, err = uc.serverRepo.ExistsWithName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check if server name exists: %w", err)
	}
	if exist {
		return nil, fmt.Errorf("server with name %s already exists", req.Name)
	}

	// Create server entity
	server := &entities.Server{
		ID:          req.ID,
		Name:        req.Name,
		Status:      entities.StatusOffline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		LastChecked: time.Now(),
		IPv4:        req.IPv4,
	}

	// Save to database
	err = uc.serverRepo.Create(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	return server, nil
}
