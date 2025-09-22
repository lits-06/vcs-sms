package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/constants"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *serverHandler) RegisterRoutes(r *gin.Engine) {
	// Swagger documentation routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Protected routes
	r.Use(h.middleware.RequireAuth())

	r.POST("/servers", h.middleware.RequireScopes(constants.ServerScopeCreate), h.CreateServer)
	r.GET("/servers", h.middleware.RequireScopes(constants.ServerScopeView), h.ViewServer)
	r.PUT("/servers", h.middleware.RequireScopes(constants.ServerScopeUpdate), h.UpdateServer)
	r.DELETE("/servers/:id", h.middleware.RequireScopes(constants.ServerScopeDelete), h.DeleteServer)
	r.POST("/servers/import", h.middleware.RequireScopes(constants.ServerScopeImport), h.ImportServersFromExcel)
	r.GET("/servers/export", h.middleware.RequireScopes(constants.ServerScopeExport), h.ExportServersToExcel)
}
