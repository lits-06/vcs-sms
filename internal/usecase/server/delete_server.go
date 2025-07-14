package server

import (
	"context"
	"fmt"

	"github.com/lits-06/vcs-sms/internal/domain/repositories"
)

type DeleteServerUseCase struct {
	serverRepo repositories.ServerRepository
}

func NewDeleteServerUseCase(serverRepo repositories.ServerRepository) *DeleteServerUseCase {
	return &DeleteServerUseCase{
		serverRepo: serverRepo,
	}
}

func (uc *DeleteServerUseCase) Execute(ctx context.Context, serverID string) error {
	if serverID == "" {
		return fmt.Errorf("server ID cannot be empty")
	}

	// Check if server exists
	exist, err := uc.serverRepo.ExistsWithID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}
	if !exist {
		return fmt.Errorf("server with ID %s does not exist", serverID)
	}

	// Delete the server
	err = uc.serverRepo.Delete(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	return nil
}
