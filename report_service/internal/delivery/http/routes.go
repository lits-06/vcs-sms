package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/constants"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *reportHandler) RegisterRoutes(r *gin.Engine) {
	// Swagger documentation routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// API routes with authentication
	r.Use(h.middleware.RequireAuth())
	r.POST("/reports", h.middleware.RequireScopes(constants.ServerScopeReport), h.CreateReport)
}
