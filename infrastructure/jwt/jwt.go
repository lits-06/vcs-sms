package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/entity"
)

type jwtAuthService struct {
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
}

func NewJWTAuthService(cfg *config.JWTConfig) *jwtAuthService {
	return &jwtAuthService{
		accessSecret:  cfg.AccessSecret,
		refreshSecret: cfg.RefreshSecret,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
		issuer:        cfg.Issuer,
	}
}

func (j *jwtAuthService) CreateCredential(user *entity.User) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         user.ID,
		"exp":         now.Add(j.accessTTL).Unix(),
		"iss":         j.issuer,
		"iat":         now.Unix(),
		"nbf":         now.Unix(),
		"jti":         uuid.New().String(),
		"role":        user.Role,
		"scopes":      user.Scopes,
		"token_types": "access",
	})

	return token.SignedString([]byte(j.accessSecret))
}

// ValidateAccess(credentialStr string) (string, error)
//
//	RefreshCredential(refreshStr string) (string, error)
//	Revoke(credentialStr string) error
func (j *jwtAuthService) ValidateAccess(credentialStr string) (string, error) {
	token, err := jwt.Parse(credentialStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(j.accessSecret), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	return "", jwt.ErrInvalidKeyType
}
