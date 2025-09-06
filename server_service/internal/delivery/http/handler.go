package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/opentracing/opentracing-go"
)

type serverHandler struct {
	log           logger.Logger
	serverUsecase domain.UseCase
}

func NewServerHandler(log logger.Logger, serverUsecase domain.UseCase) *serverHandler {
	return &serverHandler{
		log:           log,
		serverUsecase: serverUsecase,
	}
}

func (h *serverHandler) CreateServer(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "serverHandler.CreateServer")
	defer span.Finish()

	var req domain.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	server, err := h.serverUsecase.CreateServer(ctx, &req)
	if err != nil {
		h.log.Errorf("Failed to create server: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}
	h.log.Info("Server created successfully", "server_id", server.ID)
	c.JSON(http.StatusCreated, server)
}

func (h *serverHandler) ViewServer(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "serverHandler.ViewServer")
	defer span.Finish()

	var req domain.QueryServerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Errorf("Failed to bind query: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	response, err := h.serverUsecase.ViewServer(ctx, &req)
	if err != nil {
		h.log.Errorf("Failed to view server: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to view server"})
		return
	}
	h.log.Info("Server viewed successfully")
	c.JSON(http.StatusOK, response)
}

func (h *serverHandler) UpdateServer(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "serverHandler.UpdateServer")
	defer span.Finish()

	var req domain.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.serverUsecase.UpdateServer(ctx, &req); err != nil {
		h.log.Errorf("Failed to update server id=%s: %v", req.ID, tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}
	h.log.Info("Server updated successfully", "server_id", req.ID)
	c.Status(http.StatusNoContent)
}

func (h *serverHandler) DeleteServer(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "serverHandler.DeleteServer")
	defer span.Finish()

	serverID := c.Param("id")
	if err := h.serverUsecase.DeleteServer(ctx, serverID); err != nil {
		h.log.Errorf("Failed to delete server id=%s: %v", serverID, tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}
	h.log.Info("Server deleted successfully", "server_id", serverID)
	c.Status(http.StatusNoContent)
}

func (h *serverHandler) ImportServersFromExcel(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "serverHandler.ImportServersFromExcel")
	defer span.Finish()

	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		h.log.Errorf("Failed to get file from form: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.log.Errorf("Failed to open file: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	response, err := h.serverUsecase.ImportServersFromExcel(ctx, file)
	if err != nil {
		h.log.Errorf("Failed to import servers from Excel: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import servers"})
		return
	}

	h.log.Info("Servers imported successfully",
		"filename", fileHeader.Filename,
		"success_count", response.SuccessCount,
		"failure_count", response.FailureCount)
	c.JSON(http.StatusOK, response)
}

func (h *serverHandler) ExportServersToExcel(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "serverHandler.ExportServersToExcel")
	defer span.Finish()

	var req domain.QueryServerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Errorf("Failed to bind query: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	buffer, err := h.serverUsecase.ExportServersToExcel(ctx, &req)
	if err != nil {
		h.log.Errorf("Failed to export servers to Excel: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export servers"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "servers.xlsx"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Length", fmt.Sprintf("%d", len(buffer)))

	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer)
	h.log.Info("Servers exported successfully")
}

