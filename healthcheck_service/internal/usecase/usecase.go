package usecase

import (
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/domain"
)

type healthCheckUseCase struct {
	repo domain.Repository
}

func NewHealthCheckUseCase(repo domain.Repository) domain.UseCase {
	return &healthCheckUseCase{
		repo: repo,
	}
}
