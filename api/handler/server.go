package handler

import (
	"net/http"
	"path/filepath"

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

func (h *ServerHandler) ViewServer(c *gin.Context) {
	var req server.QueryServerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Failed to bind query", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	response, err := h.service.ViewServer(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to view server", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to view server"})
		return
	}
	h.logger.Info("Server viewed successfully")
	c.JSON(http.StatusOK, response)
}

func (h *ServerHandler) UpdateServer(c *gin.Context) {
	var req server.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.UpdateServer(c.Request.Context(), req); err != nil {
		h.logger.Error("Failed to update server", "server_id", req.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}
	h.logger.Info("Server updated successfully", "server_id", req.ID)
	c.Status(http.StatusNoContent)
}

func (h *ServerHandler) DeleteServer(c *gin.Context) {
	serverID := c.Param("id")
	if err := h.service.DeleteServer(c.Request.Context(), serverID); err != nil {
		h.logger.Error("Failed to delete server", "server_id", serverID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}
	h.logger.Info("Server deleted successfully", "server_id", serverID)
	c.Status(http.StatusNoContent)
}

func (h *ServerHandler) ImportServersFromExcel(c *gin.Context) {
	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get file from form", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Validate file type
	if !isValidExcelFile(fileHeader.Filename) {
		h.logger.Error("Invalid file type", "filename", fileHeader.Filename)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only Excel files (.xlsx, .xls) are allowed"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.Error("Failed to open file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	response, err := h.service.ImportServersFromExcel(c.Request.Context(), file)
	if err != nil {
		h.logger.Error("Failed to import servers from Excel", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import servers"})
		return
	}

	h.logger.Info("Servers imported successfully",
		"filename", fileHeader.Filename,
		"success_count", response.SuccessCount,
		"failure_count", response.FailureCount)
	c.JSON(http.StatusOK, response)
}

// Helper functions
func isValidExcelFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".xlsx" || ext == ".xls"
}

func (h *ServerHandler) ExportServersToExcel(c *gin.Context) {
	var req server.QueryServerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Failed to bind query", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.ExportServersToExcel(c.Request.Context(), req); err != nil {
		h.logger.Error("Failed to export servers to Excel", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export servers"})
		return
	}

	h.logger.Info("Servers exported successfully")
	c.Status(http.StatusNoContent)
}
