package domain

import "context"

type UseCase interface {
	Register(ctx context.Context, req *RegisterRequest) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	AddUserScope(ctx context.Context, userID string, scopes []string) error
	RemoveUserScope(ctx context.Context, userID string, scopes []string) error
}
