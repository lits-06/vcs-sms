package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/constants"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *reportHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/health", h.Health)

	report := r.Group("/api/reports")
	report.Use(h.middleware.RequireAuth())
	report.POST("/", h.middleware.RequireScopes(constants.ServerScopeReport), h.CreateReport)
}
