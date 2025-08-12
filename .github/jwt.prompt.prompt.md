---
description: New prompt created from chat session
mode: ask
model: Claude Sonnet 4
---
<user>
@workspace triển khai jwt cho hệ thống, yêu cầu tuân thủ clean architect và có khả năng mở rộng auth service trong tương lai
</user>
<assistant>
Dựa trên cấu trúc của dự án và yêu cầu Clean Architecture, tôi sẽ triển khai JWT system với khả năng mở rộng. Hệ thống hiện tại đã có `entity.AuthService` interface và một số component JWT cơ bản.

## 1. Cập nhật Entity Layer

````go
package entity

import "time"

// AuthService interface for authentication operations
type AuthService interface {
	CreateCredential(userID string) (string, error)
	ValidateAccess(credentialStr string) (string, error)
	RefreshCredential(refreshStr string) (string, error)
	Revoke(credentialStr string) error
}

// JWTService interface for JWT operations
type JWTService interface {
	GenerateAccessToken(userID, email string, scopes []string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateAccessToken(token string) (*JWTClaims, error)
	ValidateRefreshToken(token string) (*JWTClaims, error)
	HasScope(claims *JWTClaims, requiredScope string) bool
	ExtractClaims(token string) (*JWTClaims, error)
}

// TokenRepository interface for token storage operations
type TokenRepository interface {
	StoreRefreshToken(userID, tokenID string, expiresAt time.Time) error
	IsRefreshTokenValid(userID, tokenID string) (bool, error)
	RevokeRefreshToken(userID, tokenID string) error
	RevokeAllUserTokens(userID string) error
	CleanupExpiredTokens() error
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email,omitempty"`
	Scopes    []string  `json:"scopes"`
	TokenID   string    `json:"token_id,omitempty"`
	TokenType string    `json:"token_type"` // "access" or "refresh"
	IssuedAt  time.Time `json:"iat"`
	ExpiresAt time.Time `json:"exp"`
	Issuer    string    `json:"iss"`
}

// Credential represents authentication credential response
type Credential struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	Scope        []string  `json:"scope,omitempty"`
	IssuedAt     time.Time `json:"issued_at"`
}

// Auth scopes constants
const (
	ScopeUserRead   = "user:read"
	ScopeUserWrite  = "user:write"
	ScopeServerRead = "server:read" 
	ScopeServerWrite = "server:write"
	ScopeAdminAll   = "admin:all"
)
````

## 2. Infrastructure Layer - JWT Implementation

````go
package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lits-06/vcs-sms/entity"
)

type jwtService struct {
	accessSecret   string
	refreshSecret  string
	accessTTL      time.Duration
	refreshTTL     time.Duration
	issuer         string
}

func NewJWTService(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration, issuer string) entity.JWTService {
	return &jwtService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		issuer:        issuer,
	}
}

func (j *jwtService) GenerateAccessToken(userID, email string, scopes []string) (string, error) {
	now := time.Now()
	claims := &entity.JWTClaims{
		UserID:    userID,
		Email:     email,
		Scopes:    scopes,
		TokenType: "access",
		IssuedAt:  now,
		ExpiresAt: now.Add(j.accessTTL),
		Issuer:    j.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    claims.UserID,
		"email":      claims.Email,
		"scopes":     claims.Scopes,
		"token_type": claims.TokenType,
		"iat":        claims.IssuedAt.Unix(),
		"exp":        claims.ExpiresAt.Unix(),
		"iss":        claims.Issuer,
	})

	return token.SignedString([]byte(j.accessSecret))
}

func (j *jwtService) GenerateRefreshToken(userID string) (string, error) {
	tokenID, err := generateTokenID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := &entity.JWTClaims{
		UserID:    userID,
		TokenID:   tokenID,
		TokenType: "refresh",
		IssuedAt:  now,
		ExpiresAt: now.Add(j.refreshTTL),
		Issuer:    j.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    claims.UserID,
		"token_id":   claims.TokenID,
		"token_type": claims.TokenType,
		"iat":        claims.IssuedAt.Unix(),
		"exp":        claims.ExpiresAt.Unix(),
		"iss":        claims.Issuer,
	})

	return token.SignedString([]byte(j.refreshSecret))
}

func (j *jwtService) ValidateAccessToken(tokenString string) (*entity.JWTClaims, error) {
	return j.validateToken(tokenString, j.accessSecret, "access")
}

func (j *jwtService) ValidateRefreshToken(tokenString string) (*entity.JWTClaims, error) {
	return j.validateToken(tokenString, j.refreshSecret, "refresh")
}

func (j *jwtService) validateToken(tokenString, secret, expectedType string) (*entity.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != expectedType {
		return nil, errors.New("invalid token type")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user_id in token")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid exp in token")
	}

	if time.Now().Unix() > int64(exp) {
		return nil, errors.New("token expired")
	}

	jwtClaims := &entity.JWTClaims{
		UserID:    userID,
		TokenType: tokenType,
		ExpiresAt: time.Unix(int64(exp), 0),
		Issuer:    claims["iss"].(string),
	}

	if email, ok := claims["email"].(string); ok {
		jwtClaims.Email = email
	}

	if tokenID, ok := claims["token_id"].(string); ok {
		jwtClaims.TokenID = tokenID
	}

	if scopes, ok := claims["scopes"].([]interface{}); ok {
		scopeStrings := make([]string, len(scopes))
		for i, scope := range scopes {
			if scopeStr, ok := scope.(string); ok {
				scopeStrings[i] = scopeStr
			}
		}
		jwtClaims.Scopes = scopeStrings
	}

	return jwtClaims, nil
}

func (j *jwtService) HasScope(claims *entity.JWTClaims, requiredScope string) bool {
	for _, scope := range claims.Scopes {
		if scope == entity.ScopeAdminAll || scope == requiredScope {
			return true
		}
	}
	return false
}

func (j *jwtService) ExtractClaims(tokenString string) (*entity.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.accessSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	jwtClaims := &entity.JWTClaims{}
	if userID, ok := claims["user_id"].(string); ok {
		jwtClaims.UserID = userID
	}
	if email, ok := claims["email"].(string); ok {
		jwtClaims.Email = email
	}

	return jwtClaims, nil
}

func generateTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
````

## 3. Infrastructure Layer - Redis Token Repository

````go
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/lits-06/vcs-sms/entity"
	"github.com/redis/go-redis/v9"
)

const (
	RefreshTokenKey = "refresh_token:%s:%s" // refresh_token:{user_id}:{token_id}
	UserTokensKey   = "user_tokens:%s"      // user_tokens:{user_id} - set of token_ids
)

type tokenRepository struct {
	client *redis.Client
}

func NewTokenRepository(client *redis.Client) entity.TokenRepository {
	return &tokenRepository{
		client: client,
	}
}

func (r *tokenRepository) StoreRefreshToken(userID, tokenID string, expiresAt time.Time) error {
	ctx := context.Background()
	
	// Store refresh token with expiration
	key := fmt.Sprintf(RefreshTokenKey, userID, tokenID)
	ttl := time.Until(expiresAt)
	if err := r.client.Set(ctx, key, "valid", ttl).Err(); err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Add token to user's token set
	userTokensKey := fmt.Sprintf(UserTokensKey, userID)
	if err := r.client.SAdd(ctx, userTokensKey, tokenID).Err(); err != nil {
		return fmt.Errorf("failed to add token to user set: %w", err)
	}

	// Set expiration for user tokens set
	if err := r.client.Expire(ctx, userTokensKey, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set expiration for user tokens: %w", err)
	}

	return nil
}

func (r *tokenRepository) IsRefreshTokenValid(userID, tokenID string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf(RefreshTokenKey, userID, tokenID)
	
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check refresh token: %w", err)
	}

	return result > 0, nil
}

func (r *tokenRepository) RevokeRefreshToken(userID, tokenID string) error {
	ctx := context.Background()
	
	// Remove refresh token
	key := fmt.Sprintf(RefreshTokenKey, userID, tokenID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	// Remove from user's token set
	userTokensKey := fmt.Sprintf(UserTokensKey, userID)
	if err := r.client.SRem(ctx, userTokensKey, tokenID).Err(); err != nil {
		return fmt.Errorf("failed to remove token from user set: %w", err)
	}

	return nil
}

func (r *tokenRepository) RevokeAllUserTokens(userID string) error {
	ctx := context.Background()
	
	// Get all token IDs for user
	userTokensKey := fmt.Sprintf(UserTokensKey, userID)
	tokenIDs, err := r.client.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Remove all refresh tokens
	pipe := r.client.Pipeline()
	for _, tokenID := range tokenIDs {
		key := fmt.Sprintf(RefreshTokenKey, userID, tokenID)
		pipe.Del(ctx, key)
	}
	
	// Remove user tokens set
	pipe.Del(ctx, userTokensKey)
	
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	return nil
}

func (r *tokenRepository) CleanupExpiredTokens() error {
	// Redis automatically handles expiration, so this can be a no-op
	// or implement scanning and cleanup logic if needed
	return nil
}
````

## 4. Services Layer - Auth Service Implementation

````go
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lits-06/vcs-sms/entity"
)

type AuthService struct {
	jwtService    entity.JWTService
	tokenRepo     entity.TokenRepository
	userService   UserService // Interface for user operations
}

type UserService interface {
	GetUserByID(userID string) (*entity.User, error)
	GetUserScopes(userID string) ([]string, error)
}

func NewAuthService(
	jwtService entity.JWTService, 
	tokenRepo entity.TokenRepository,
	userService UserService,
) entity.AuthService {
	return &AuthService{
		jwtService:  jwtService,
		tokenRepo:   tokenRepo,
		userService: userService,
	}
}

func (s *AuthService) CreateCredential(userID string) (string, error) {
	// Get user information
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Get user scopes
	scopes, err := s.userService.GetUserScopes(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user scopes: %w", err)
	}

	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(userID, user.Email, scopes)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.jwtService.GenerateRefreshToken(userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Extract token ID from refresh token for storage
	refreshClaims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.tokenRepo.StoreRefreshToken(userID, refreshClaims.TokenID, refreshClaims.ExpiresAt); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create credential response
	credential := &entity.Credential{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(refreshClaims.ExpiresAt).Seconds()),
		Scope:        scopes,
		IssuedAt:     time.Now(),
	}

	// Serialize credential
	credentialBytes, err := json.Marshal(credential)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential: %w", err)
	}

	return string(credentialBytes), nil
}

func (s *AuthService) ValidateAccess(credentialStr string) (string, error) {
	var credential entity.Credential
	if err := json.Unmarshal([]byte(credentialStr), &credential); err != nil {
		return "", fmt.Errorf("failed to unmarshal credential: %w", err)
	}

	// Validate access token
	claims, err := s.jwtService.ValidateAccessToken(credential.AccessToken)
	if err != nil {
		return "", fmt.Errorf("invalid access token: %w", err)
	}

	return claims.UserID, nil
}

func (s *AuthService) RefreshCredential(refreshStr string) (string, error) {
	// Parse refresh token from credential
	var credential entity.Credential
	if err := json.Unmarshal([]byte(refreshStr), &credential); err != nil {
		return "", fmt.Errorf("failed to unmarshal credential: %w", err)
	}

	// Validate refresh token
	refreshClaims, err := s.jwtService.ValidateRefreshToken(credential.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if refresh token exists in storage
	isValid, err := s.tokenRepo.IsRefreshTokenValid(refreshClaims.UserID, refreshClaims.TokenID)
	if err != nil {
		return "", fmt.Errorf("failed to check refresh token: %w", err)
	}

	if !isValid {
		return "", errors.New("refresh token has been revoked")
	}

	// Generate new credential
	newCredential, err := s.CreateCredential(refreshClaims.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to create new credential: %w", err)
	}

	// Revoke old refresh token
	if err := s.tokenRepo.RevokeRefreshToken(refreshClaims.UserID, refreshClaims.TokenID); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to revoke old refresh token: %v\n", err)
	}

	return newCredential, nil
}

func (s *AuthService) Revoke(credentialStr string) error {
	var credential entity.Credential
	if err := json.Unmarshal([]byte(credentialStr), &credential); err != nil {
		return fmt.Errorf("failed to unmarshal credential: %w", err)
	}

	// Validate refresh token to get claims
	refreshClaims, err := s.jwtService.ValidateRefreshToken(credential.RefreshToken)
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}

	// Revoke refresh token
	if err := s.tokenRepo.RevokeRefreshToken(refreshClaims.UserID, refreshClaims.TokenID); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}
````

## 5. Cập nhật Config

````go
// ...existing code...

type JWTConfig struct {
	AccessSecret   string        `yaml:"access_secret" env:"JWT_ACCESS_SECRET"`
	RefreshSecret  string        `yaml:"refresh_secret" env:"JWT_REFRESH_SECRET"`
	AccessTTL      time.Duration `yaml:"access_ttl" env:"JWT_ACCESS_TTL" envDefault:"15m"`
	RefreshTTL     time.Duration `yaml:"refresh_ttl" env:"JWT_REFRESH_TTL" envDefault:"168h"`
	Issuer         string        `yaml:"issuer" env:"JWT_ISSUER" envDefault:"vcs-sms"`
}

type Config struct {
	// ...existing code...
	JWT JWTConfig `yaml:"jwt"`
}
````

## 6. API Middleware cho Authorization

````go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/entity"
)

type AuthMiddleware struct {
	jwtService entity.JWTService
}

func NewAuthMiddleware(jwtService entity.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_scopes", claims.Scopes)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireScope(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !m.jwtService.HasScope(claims, requiredScope) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_scopes", claims.Scopes)
		c.Next()
	}
}

func extractTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// Helper function to get user ID from context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}
````

Hệ thống JWT này tuân thủ Clean Architecture với:

1. **Entity Layer**: Định nghĩa interfaces và structs cốt lõi
2. **Infrastructure Layer**: Triển khai cụ thể cho JWT và Redis storage
3. **Services Layer**: Business logic cho authentication
4. **API Layer**: Middleware cho authorization

**Khả năng mở rộng**:
- Dễ dàng thêm providers mới (OAuth, SAML, etc.)
- Support multiple token storage backends
- Flexible scope system
- Extensible middleware system

**Features chính**:
- Access/Refresh token pattern
- Scope-based authorization
- Token revocation
- Clean separation of concerns
- Redis-based token storage
- Comprehensive error handling
</assistant>