package usecase

import (
	"context"
	"fmt"

	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/pkg/utils"
	"github.com/lits-06/vcs-sms/user_service/internal/domain"
	"github.com/opentracing/opentracing-go"
)

type userUsecase struct {
	userRepo domain.Repository
}

func NewUserUsecase(userRepo domain.Repository) domain.UseCase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) Register(ctx context.Context, req *domain.RegisterRequest) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUsecase.Register")
	defer span.Finish()

	existUser, err := u.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to get user by email: %w", err))
	}

	if existUser != nil {
		return tracing.TraceWithErr(span, domain.ErrUserExists)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to hash password: %w", err))
	}

	user := &domain.User{
		Email:    req.Email,
		Username: req.Username,
		Password: hashedPassword,
		Scopes:   domain.DefaultScopes(),
	}

	err = u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to create user: %w", err))
	}

	return nil
}

func (u *userUsecase) AddUserScope(ctx context.Context, userID string, scopes []string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUsecase.AddUserScope")
	defer span.Finish()

	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to get user by ID: %w", err))
	}
	if user == nil {
		return tracing.TraceWithErr(span, domain.ErrUserNotFound)
	}

	for _, scope := range scopes {
		if !domain.IsValidScope(scope) {
			return tracing.TraceWithErr(span, domain.ErrInvalidScope)
		}
	}

	err = u.userRepo.AddUserScopes(ctx, userID, scopes)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to add user scopes: %w", err))
	}

	return nil
}

func (u *userUsecase) RemoveUserScope(ctx context.Context, userID string, scopes []string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUsecase.RemoveUserScope")
	defer span.Finish()

	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to get user by ID: %w", err))
	}
	if user == nil {
		return tracing.TraceWithErr(span, domain.ErrUserNotFound)
	}

	for _, scope := range scopes {
		if !domain.IsValidScope(scope) {
			return tracing.TraceWithErr(span, domain.ErrInvalidScope)
		}
	}

	err = u.userRepo.RemoveUserScopes(ctx, userID, scopes)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to remove user scopes: %w", err))
	}

	return nil
}
