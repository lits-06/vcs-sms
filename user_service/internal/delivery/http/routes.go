package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/constants"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *userHandler) RegisterRoutes(r *gin.Engine) {
	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	r.POST("/register", h.Register)

	user := r.Group("/user")
	user.POST("/scopes", h.middleware.RequireAuth(), h.middleware.RequireScopes(constants.UserScopeUpdate), h.AddUserScope)
	user.DELETE("/scopes", h.middleware.RequireAuth(), h.middleware.RequireScopes(constants.UserScopeUpdate), h.RemoveUserScope)
}
