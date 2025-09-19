package repository

import (
	"context"
	"fmt"

	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/user_service/internal/domain"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.Repository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userRepository.CreateUser")
	defer span.Finish()

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to create user: %w", err))
	}

	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userRepository.GetUserByEmail")
	defer span.Finish()

	var user domain.User
	if err := r.db.WithContext(ctx).Preload("Scopes").Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to get user by email: %w", err))
	}

	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userRepository.GetUserByID")
	defer span.Finish()

	var user domain.User
	if err := r.db.WithContext(ctx).Preload("Scopes").Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to get user by ID: %w", err))
	}

	return &user, nil
}

func (r *userRepository) AddUserScopes(ctx context.Context, userID string, scopes []string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userRepository.AddUserScopes")
	defer span.Finish()

	var user domain.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to find user: %w", err))
	}

	var sc []domain.Scope
	if err := r.db.WithContext(ctx).Where("name IN ?", scopes).Find(&sc).Error; err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to find scopes: %w", err))
	}
	if len(sc) == 0 {
		return tracing.TraceWithErr(span, fmt.Errorf("no valid scopes found"))
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Association("Scopes").Append(sc); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to add scopes to user: %w", err))
	}

	return nil
}

func (r *userRepository) RemoveUserScopes(ctx context.Context, userID string, scopes []string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userRepository.RemoveUserScopes")
	defer span.Finish()

	var user domain.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to find user: %w", err))
	}

	var sc []domain.Scope
	if err := r.db.WithContext(ctx).Where("name IN ?", scopes).Find(&sc).Error; err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to find scopes: %w", err))
	}
	if len(sc) == 0 {
		return tracing.TraceWithErr(span, fmt.Errorf("no valid scopes found"))
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Association("Scopes").Delete(sc); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to remove scopes from user: %w", err))
	}

	return nil
}
