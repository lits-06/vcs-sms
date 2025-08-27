package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateToken(userID, email string, scopes []string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	HasScope(claims *Claims, requiredScope string) bool
}

const (
	// Server scopes
	ScopeServerRead   = "server:read"
	ScopeServerWrite  = "server:write"
	ScopeServerDelete = "server:delete"
	ScopeServerUpdate = "server:update"
	ScopeServerImport = "server:import"
	ScopeServerExport = "server:export"

	// User scopes
	ScopeUserRead   = "user:read"
	ScopeUserWrite  = "user:write"
	ScopeUserDelete = "user:delete"
	ScopeUserUpdate = "user:update"

	// Admin scopes
	ScopeAdminAll = "admin:all"
)

type jwtService struct {
	secretKey string
	issuer    string
}

func NewJWTService(secretKey, issuer string) JWTService {
	return &jwtService{
		secretKey: secretKey,
		issuer:    issuer,
	}
}

func (j *jwtService) GenerateToken(userID, email string, scopes []string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *jwtService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (j *jwtService) HasScope(claims *Claims, requiredScope string) bool {
	// Admin có tất cả quyền
	for _, scope := range claims.Scopes {
		if scope == ScopeAdminAll {
			return true
		}
		if scope == requiredScope {
			return true
		}
	}
	return false
}
