package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/usecases/server"
)

type ServerHandler struct {
	service server.UseCase
	logger  logger.Logger
}

func NewServerHandler(service server.UseCase, logger logger.Logger) *ServerHandler {
	log := logger.With("handler", "server")

	return &ServerHandler{
		service: service,
		logger:  log,
	}
}

func (h *ServerHandler) CreateServer(c *gin.Context) {
	var req server.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	server, err := h.service.CreateServer(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create server", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}
	h.logger.Info("Server created successfully", "server_id", server.ID)
	c.JSON(http.StatusCreated, server)
}

// func (h *ServerHandler) UpdateServer(c *gin.Context) {
// 	id := c.Param("id")
// 	var req server.UpdateServerRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		h.logger.Error("Failed to bind JSON", "error", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
// 		return
// 	}

// 	serverID := c.Param("id")
// 	if err := h.service.UpdateServer(c.Request.Context(), req); err != nil {
// 		h.logger.Error("Failed to update server", "server_id", serverID, "error", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
// 		return
// 	}
// 	h.logger.Info("Server updated successfully", "server_id", serverID)
// 	c.Status(http.StatusNoContent)
// }
