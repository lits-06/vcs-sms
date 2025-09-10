package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lits-06/vcs-sms/auth_service/config"
	"github.com/lits-06/vcs-sms/auth_service/internal/domain"
	"github.com/lits-06/vcs-sms/auth_service/internal/dto"
	"github.com/lits-06/vcs-sms/pkg-old/utils"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	userpb "github.com/lits-06/vcs-sms/proto"
	"github.com/opentracing/opentracing-go"
)

type authUsecase struct {
	cfg       *config.Config
	cacheRepo domain.CacheRepository
	usClient  userpb.UserServiceClient
}

func NewAuthUsecase(cfg *config.Config, cacheRepo domain.CacheRepository, usClient userpb.UserServiceClient) domain.UseCase {
	return &authUsecase{
		cfg:       cfg,
		cacheRepo: cacheRepo,
		usClient:  usClient,
	}
}

func (a *authUsecase) Login(ctx context.Context, email, password string) (string, string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "authUsecase.Login")
	defer span.Finish()

	res, err := a.usClient.GetUserByEmail(ctx, &userpb.GetUserByEmailRequest{Email: email})
	if err != nil {
		return "", "", tracing.TraceWithErr(span, fmt.Errorf("usClient.GetUserByEmail: %w", err))
	}

	user := dto.UserResponseFromGrpc(res.User)

	hashedPassword := user.Password
	err = utils.CheckPasswordHash(password, hashedPassword)
	if err != nil {
		return "", "", tracing.TraceWithErr(span, fmt.Errorf("utils.CheckPasswordHash: %w", err))
	}

	scopes := domain.ScopesToStringSlice(user.Scopes)

	accessToken, err := a.GenerateAccessToken(ctx, email, scopes)
	if err != nil {
		return "", "", tracing.TraceWithErr(span, fmt.Errorf("GenerateAccessToken: %w", err))
	}

	refreshToken, err := a.GenerateRefreshToken(ctx)
	if err != nil {
		return "", "", tracing.TraceWithErr(span, fmt.Errorf("GenerateRefreshToken: %w", err))
	}

	err = a.cacheRepo.SetRefreshToken(ctx, refreshToken, user.Email, a.cfg.JWT.RefreshTTL)
	if err != nil {
		return "", "", tracing.TraceWithErr(span, fmt.Errorf("cacheRepo.Set: %w", err))
	}

	return accessToken, refreshToken, nil
}

func (a *authUsecase) Logout(ctx context.Context, token string) error {
	// Implement logout logic here
	return nil
}

func (a *authUsecase) GenerateAccessToken(ctx context.Context, email string, scopes []string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "authUsecase.GenerateAccessToken")
	defer span.Finish()

	accessClaim := &domain.AccessClaim{
		Email:  email,
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.cfg.JWT.AccessTTL)),
		},
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaim)
	accessToken, err := access.SignedString([]byte(a.cfg.JWT.AccessSecretKey))
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("access.SignedString: %w", err))
	}

	return accessToken, nil
}

func (a *authUsecase) GenerateRefreshToken(ctx context.Context) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "authUsecase.GenerateRefreshToken")
	defer span.Finish()

	refreshClaim := &domain.RefreshClaim{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.cfg.JWT.RefreshTTL)),
		},
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaim)
	refreshToken, err := refresh.SignedString([]byte(a.cfg.JWT.RefreshSecretKey))
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("refresh.SignedString: %w", err))
	}

	return refreshToken, nil
}

func (a *authUsecase) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "authUsecase.RefreshAccessToken")
	defer span.Finish()

	// Parse and validate the refresh token
	token, err := jwt.ParseWithClaims(refreshToken, &domain.RefreshClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.cfg.JWT.RefreshSecretKey), nil
	})
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("jwt.ParseWithClaims: %w", err))
	}

	_, ok := token.Claims.(*domain.RefreshClaim)
	if !ok || !token.Valid {
		return "", tracing.TraceWithErr(span, fmt.Errorf("invalid token claims"))
	}

	// Check if the refresh token is blacklisted
	isBlacklisted, err := a.cacheRepo.IsBlacklisted(ctx, refreshToken)
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("cacheRepo.IsBlacklisted: %w", err))
	}
	if isBlacklisted {
		return "", tracing.TraceWithErr(span, fmt.Errorf("refresh token is invalid"))
	}

	// Get the email associated with the refresh token
	email, err := a.cacheRepo.GetEmailByRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("cacheRepo.GetEmailByRefreshToken: %w", err))
	}

	// Here you might want to fetch user scopes from user service again
	res, err := a.usClient.GetUserByEmail(ctx, &userpb.GetUserByEmailRequest{Email: email})
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("usClient.GetUserByEmail: %w", err))
	}
	user := dto.UserResponseFromGrpc(res.User)
	scopes := domain.ScopesToStringSlice(user.Scopes)

	// Generate a new access token
	newAccessToken, err := a.GenerateAccessToken(ctx, email, scopes)
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("GenerateAccessToken: %w", err))
	}

	return newAccessToken, nil
}

func (a *authUsecase) RevokeToken(ctx context.Context, accessToken string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "authUsecase.RevokeToken")
	defer span.Finish()

	// Parse the token to get its expiration time
	parsedToken, err := jwt.ParseWithClaims(accessToken, &domain.AccessClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.cfg.JWT.AccessSecretKey), nil
	})
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("jwt.ParseWithClaims: %w", err))
	}

	claims, ok := parsedToken.Claims.(*domain.AccessClaim)
	if !ok || !parsedToken.Valid {
		return tracing.TraceWithErr(span, fmt.Errorf("invalid token claims"))
	}

	expiration := claims.ExpiresAt.Unix()

	// Add the token to the blacklist
	err = a.cacheRepo.AddBlacklist(ctx, accessToken, expiration)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("cacheRepo.AddBlacklist: %w", err))
	}

	return nil
}

func (a *authUsecase) RevokeUserTokens(ctx context.Context, userID string) error {
	// Implement logic to revoke all tokens for a specific user

	return nil
}
