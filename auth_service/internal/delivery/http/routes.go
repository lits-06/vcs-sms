package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *authHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/health", h.Health)

	auth := r.Group("/api/auth")
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.RefreshAccessToken)
}
