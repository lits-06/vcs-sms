package domain

import "context"

type UseCase interface {
	Register(ctx context.Context, email, username, password string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	AddUserScope(ctx context.Context, email string, scopes []string) error
	RemoveUserScope(ctx context.Context, email string, scopes []string) error
}
