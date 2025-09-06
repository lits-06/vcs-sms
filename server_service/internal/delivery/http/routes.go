package http

import "github.com/gin-gonic/gin"

func (h *serverHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/servers", h.CreateServer)
	r.GET("/servers", h.ViewServer)
	r.PUT("/servers", h.UpdateServer)
	r.DELETE("/servers/:id", h.DeleteServer)
	r.POST("/servers/import", h.ImportServersFromExcel)
	r.GET("/servers/export", h.ExportServersToExcel)
}
