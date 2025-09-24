package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/constants"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *userHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/health", h.Health)
	user := r.Group("/api/users")
	user.POST("/register", h.Register)
	user.Use(h.middleware.RequireAuth())
	user.POST("/scopes", h.middleware.RequireScopes(constants.UserScopeUpdate), h.AddUserScope)
	user.DELETE("/scopes", h.middleware.RequireScopes(constants.UserScopeUpdate), h.RemoveUserScope)
}
