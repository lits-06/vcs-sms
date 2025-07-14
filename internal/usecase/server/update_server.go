package server

import (
	"context"
	"fmt"

	"github.com/lits-06/vcs-sms/internal/domain/repositories"
)

type UpdateServerUseCase struct {
	serverRepo repositories.ServerRepository
}

func NewUpdateServerUseCase(serverRepo repositories.ServerRepository) *UpdateServerUseCase {
	return &UpdateServerUseCase{
		serverRepo: serverRepo,
	}
}

type UpdateServerRequest struct {
	ServerID   string                 `json:"id"`
	UpdateData map[string]interface{} `json:"update_data"`
}

func (uc *UpdateServerUseCase) Execute(ctx context.Context, req UpdateServerRequest) error {
	// Validate input
	if req.ServerID == "" {
		return nil, fmt.Errorf("server ID is required")
	}

	if len(input.UpdateData) == 0 {
		return nil, fmt.Errorf("update data is required")
	}

	// Check if server exists
	exists, err := uc.serverRepo.ExistsWithID(ctx, req.ServerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check server existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("server with ID %s does not exist", req.ServerID)
	}

	// Update server
	err = uc.serverRepo.Update(ctx, input.ServerID, input.UpdateData)
	if err != nil {
		return nil, fmt.Errorf("failed to update server: %w", err)
	}

	return nil
}
