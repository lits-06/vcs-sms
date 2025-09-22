package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/constants"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *serverHandler) RegisterRoutes(r *gin.Engine) {
	server := r.Group("/api/servers")
	server.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server Service is running",
		})
	})
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Protected routes
	server.Use(h.middleware.RequireAuth())

	server.POST("/", h.middleware.RequireScopes(constants.ServerScopeCreate), h.CreateServer)
	server.GET("/", h.middleware.RequireScopes(constants.ServerScopeView), h.ViewServer)
	server.PUT("/", h.middleware.RequireScopes(constants.ServerScopeUpdate), h.UpdateServer)
	server.DELETE("/:id", h.middleware.RequireScopes(constants.ServerScopeDelete), h.DeleteServer)
	server.POST("/import", h.middleware.RequireScopes(constants.ServerScopeImport), h.ImportServersFromExcel)
	server.GET("/export", h.middleware.RequireScopes(constants.ServerScopeExport), h.ExportServersToExcel)
}
