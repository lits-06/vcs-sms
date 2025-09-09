package domain

import "context"

type UseCase interface {
	Register(ctx context.Context, req *RegisterRequest) error
	AddUserScope(ctx context.Context, userID string, scopes []string) error
	RemoveUserScope(ctx context.Context, userID string, scopes []string) error
}
