package grpc

import (
	"context"
	"fmt"

	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	userpb "github.com/lits-06/vcs-sms/proto"
	"github.com/lits-06/vcs-sms/user_service/internal/domain"
)

type userService struct {
	userpb.UnimplementedUserServiceServer
	log    logger.Logger
	userUC domain.UseCase
}

func NewUserService(log logger.Logger, userUC domain.UseCase) *userService {
	return &userService{
		log:    log,
		userUC: userUC,
	}
}

func (s *userService) GetUserByEmail(ctx context.Context, req *userpb.GetUserByEmailRequest) (*userpb.GetUserByEmailResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "userService.GetUserByEmail")
	defer span.Finish()

	user, err := s.userUC.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.log.Error("userUC.GetUserByEmail: %v", err)
		return nil, tracing.TraceWithErr(span, fmt.Errorf("userUC.GetUserByEmail: %w", err))
	}

	return &userpb.GetUserByEmailResponse{
		User: &userpb.User{
			Id:     user.ID,
			Email:  user.Email,
			Password: user.Password,
			Scopes: convertScopes(&user.Scopes),
		},
	}, nil
}

func convertScopes(scopes *[]domain.Scope) []*userpb.Scope {
	if scopes == nil {
		return nil
	}

	result := make([]*userpb.Scope, len(*scopes))
	for i, scope := range *scopes {
		result[i] = &userpb.Scope{Name: scope.Name}
	}
	return result
}
