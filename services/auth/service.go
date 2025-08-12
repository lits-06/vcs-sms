package auth

import "github.com/lits-06/vcs-sms/entity"

type AuthService struct {
	authService entity.AuthService
}

func NewAuthService(authService entity.AuthService) *AuthService {
	return &AuthService{
		authService: authService,
	}
}
