package dto

import (
	"github.com/lits-06/vcs-sms/auth_service/internal/domain"
	userpb "github.com/lits-06/vcs-sms/proto"
)

func UserResponseFromGrpc(user *userpb.User) *domain.User {
	if user == nil {
		return nil
	}

	return &domain.User{
		ID:       user.Id,
		Email:    user.Email,
		Password: user.Password,
		Scopes:   ScopesFromGrpc(user.Scopes),
	}
}

func ScopesFromGrpc(scopes []*userpb.Scope) []domain.Scope {
	if scopes == nil {
		return nil
	}

	result := make([]domain.Scope, len(scopes))
	for i, scope := range scopes {
		result[i] = domain.Scope{Name: scope.Name}
	}
	return result
}
