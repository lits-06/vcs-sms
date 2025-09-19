package domain

import "context"

type Repository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	AddUserScopes(ctx context.Context, userID string, scopes []string) error
	RemoveUserScopes(ctx context.Context, userID string, scopes []string) error
}
