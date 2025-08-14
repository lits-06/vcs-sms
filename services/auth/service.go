package auth

import (
	"errors"

	"github.com/lits-06/vcs-sms/domain/entity"
	"github.com/lits-06/vcs-sms/domain/repository"
	"github.com/lits-06/vcs-sms/pkg/utils"
)

type authService struct {
	tokenProvider         entity.TokenProvider
	authorizationProvider entity.AuthorizationProvider
	userRepo              repository.UserRepository
}

func NewAuthService(tokenProvider entity.TokenProvider, authorizationProvider entity.AuthorizationProvider, userRepo repository.UserRepository) *authService {
	return &authService{
		tokenProvider:         tokenProvider,
		authorizationProvider: authorizationProvider,
		userRepo:              userRepo,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *authService) Register(req RegisterRequest) (string, error) {
	// Registration logic here
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return "", err
	}

	if user != nil {
		return "", errors.New("user already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return "", err
	}

	newUser := &entity.User{
		Email:    req.Email,
		Username: req.Username,
		Password: hashedPassword,
		Role:     entity.DefaultUserRole(),   // Default role, can be changed based on requirements
		Scopes:   entity.DefaultUserScopes(), // Default scopes, can be customized
	}

	err = s.userRepo.CreateUser(newUser)
	if err != nil {
		return "", err
	}

	token, err := s.tokenProvider.GenerateToken(newUser.ID)
}

func (s *authService) Login(req LoginRequest) (string, error) {
	// Login logic here
}

func (s *authService) Logout() error {
	// Logout logic here
}
