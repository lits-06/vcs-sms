package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/constants"
)

func (h *userHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/register", h.Register)
	r.POST("/user/scopes", h.middleware.RequireAuth(), h.middleware.RequireScopes(constants.UserScopeUpdate), h.AddUserScope)
	r.DELETE("/user/scopes", h.middleware.RequireAuth(), h.middleware.RequireScopes(constants.UserScopeUpdate), h.RemoveUserScope)
}
