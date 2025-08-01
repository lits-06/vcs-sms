package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/services/server"
)

type ReportHandler struct {
	service server.UptimeService
	logger  logger.Logger
}

func NewReportHandler(service server.UptimeService, logger logger.Logger) *ReportHandler {
	log := logger.With("handler", "report")

	return &ReportHandler{
		service: service,
		logger:  log,
	}
}

func (h *ReportHandler) GenerateUptimeReport(c *gin.Context) {
	var req server.UptimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.ReportStats(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to generate uptime report", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate uptime report"})
		return
	}

	h.logger.Info("Uptime report generated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Uptime report generated successfully"})
}
