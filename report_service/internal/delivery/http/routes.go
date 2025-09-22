package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/constants"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *reportHandler) RegisterRoutes(r *gin.Engine) {
	report := r.Group("/api/reports")
	report.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	report.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Report Service is running",
		})
	})

	// API routes with authentication
	report.Use(h.middleware.RequireAuth())
	report.POST("/", h.middleware.RequireScopes(constants.ServerScopeReport), h.CreateReport)
}
