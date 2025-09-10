package domain

import (
	"context"
)

type UseCase interface {
	Login(ctx context.Context, email, password string) (string, string, error)
	Logout(ctx context.Context, token string) error
	GenerateAccessToken(ctx context.Context, email string, scopes []string) (string, error)
	GenerateRefreshToken(ctx context.Context) (string, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	RevokeToken(ctx context.Context, token string) error
	RevokeUserTokens(ctx context.Context, userID string) error
}
