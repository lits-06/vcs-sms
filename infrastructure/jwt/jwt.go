package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/entity"
)

type jwtProvider struct {
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
}

func NewJWTProvider(cfg *config.JWTConfig) *jwtProvider {
	return &jwtProvider{
		accessSecret:  cfg.AccessSecret,
		refreshSecret: cfg.RefreshSecret,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
		issuer:        cfg.Issuer,
	}
}

func (j *jwtProvider) GenerateToken(user *entity.User) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         user.ID,
		"exp":         now.Add(j.accessTTL).Unix(),
		"iss":         j.issuer,
		"iat":         now.Unix(),
		"role":        user.Role,
		"scopes":      user.Scopes,
		"token_types": "access",
	})

	return token.SignedString([]byte(j.accessSecret))
}

func (j *jwtProvider) ValidateToken(token string) (*entity.User, error) {
	token, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(j.accessSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	user := &entity.User{
		ID:     claims["sub"].(string),
		Role:   claims["role"].(string),
		Scopes: claims["scopes"].([]string),
	}

	return user, nil
}

func (j *jwtProvider) RevokeToken(token string) error {
	// Implementation for revoking a token
	return nil
}
