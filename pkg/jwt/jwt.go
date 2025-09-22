package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	SecretKey string
}

type AccessClaim struct {
	jwt.RegisteredClaims
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
}

type RefreshClaim struct {
	jwt.RegisteredClaims
}
