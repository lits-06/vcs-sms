package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	httpresponse "github.com/lits-06/vcs-sms/pkg/http_response"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/middleware"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/lits-06/vcs-sms/server_service/internal/dto"
)

type serverHandler struct {
	log           logger.Logger
	serverUsecase domain.UseCase
	middleware    *middleware.AuthMiddleware
}

func NewServerHandler(log logger.Logger, serverUsecase domain.UseCase, middleware *middleware.AuthMiddleware) *serverHandler {
	return &serverHandler{
		log:           log,
		serverUsecase: serverUsecase,
		middleware:    middleware,
	}
}

// Health checks the health status of the server service
// Health godoc
//
//	@Summary		Health check
//	@Description	Check if the Server Service is running and healthy
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httpresponse.Response	"Service is healthy"
//	@Router			/health [get]
func (h *serverHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: "Server Service is running",
	})
}

// CreateServer creates a new server
// CreateServer godoc
//
//	@Summary		Create a new server
//	@Description	Create a new server with specified name, status, IPv4 address and port
//	@Tags			Server
//	@Accept			json
//	@Produce		json
//	@Param			server	body		dto.CreateServerRequest	true	"Server data"
//	@Success		201		{object}	httpresponse.Response		"Server created successfully"
//	@Failure		400		{object}	httpresponse.Response		"Invalid request"
//	@Failure		401		{object}	httpresponse.Response		"Unauthorized"
//	@Failure		403		{object}	httpresponse.Response		"Forbidden"
//	@Failure		500		{object}	httpresponse.Response		"Failed to create server"
//	@Security		BearerAuth
//	@Router			/api/servers [post]
func (h *serverHandler) CreateServer(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "serverHandler.CreateServer")
	defer span.Finish()

	var req dto.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, httpresponse.Response{
			Message: "Invalid request",
		})
		return
	}

	server, err := h.serverUsecase.CreateServer(ctx, req.Name, req.IPv4, req.Port)
	if err != nil {
		h.log.Errorf("Failed to create server: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to create server",
		})
		return
	}
	h.log.Info("Server created successfully ", "server_id: ", server.ID)
	c.JSON(http.StatusCreated, httpresponse.Response{
		Message: "Server created successfully",
	})
}

// ViewServer retrieves servers based on query parameters
// ViewServer godoc
//
//	@Summary		View servers
//	@Description	Retrieve servers with optional filtering, sorting and pagination
//	@Tags			Server
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"Filter by server name"
//	@Param			status	query		string	false	"Filter by server status (ON/OFF)"
//	@Param			ipv4	query		string	false	"Filter by IPv4 address"
//	@Param			from	query		int		false	"Pagination offset"
//	@Param			to		query		int		false	"Pagination limit"
//	@Param			sort	query		string	false	"Sort by field (name, status, created_at, updated_at)"
//	@Param			order	query		string	false	"Sort order (asc, desc)"
//	@Success		200		{object}	httpresponse.Response		"List of servers with total count"
//	@Failure		401		{object}	httpresponse.Response		"Unauthorized"
//	@Failure		403		{object}	httpresponse.Response		"Forbidden"
//	@Failure		400		{object}	httpresponse.Response		"Invalid request"
//	@Failure		500		{object}	httpresponse.Response		"Failed to view server"
//	@Security		BearerAuth
//	@Router			/api/servers [get]
func (h *serverHandler) ViewServer(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "serverHandler.ViewServer")
	defer span.Finish()

	var req dto.QueryServerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Errorf("Failed to bind query: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, httpresponse.Response{
			Message: "Invalid request",
		})
		return
	}

	servers, total, err := h.serverUsecase.ViewServer(ctx, req.Name, req.Status, req.IPv4, req.From, req.To, req.Sort, req.Order)
	if err != nil {
		h.log.Errorf("Failed to view server: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to view server",
		})
		return
	}
	h.log.Info("Server viewed successfully ", "total_servers: ", total)
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: fmt.Sprintf("Total servers: %d", total),
		Data:    servers,
	})
}

// UpdateServer updates an existing server
// UpdateServer godoc
//
//	@Summary		Update a server
//	@Description	Update server information including name and IPv4 address
//	@Tags			Server
//	@Accept			json
//	@Produce		json
//	@Param			server	body		dto.UpdateServerRequest	true	"Server update data"
//	@Success		200		{object}	httpresponse.Response		"Server updated successfully"
//	@Failure		401		{object}	httpresponse.Response		"Unauthorized"
//	@Failure		403		{object}	httpresponse.Response		"Forbidden"
//	@Failure		400		{object}	httpresponse.Response		"Invalid request"
//	@Failure		500		{object}	httpresponse.Response		"Failed to update server"
//	@Security		BearerAuth
//	@Router			/api/servers [put]
func (h *serverHandler) UpdateServer(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "serverHandler.UpdateServer")
	defer span.Finish()

	var req dto.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, httpresponse.Response{
			Message: "Invalid request",
		})
		return
	}

	if err := h.serverUsecase.UpdateServer(ctx, req.ID, req.Name, req.IPv4); err != nil {
		h.log.Errorf("Failed to update server id=%s: %v", req.ID, tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to update server",
		})
		return
	}
	h.log.Info("Server updated successfully ", "server_id: ", req.ID)
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: "Server updated successfully",
	})
}

// DeleteServer deletes a server by ID
// DeleteServer godoc
//
//	@Summary		Delete a server
//	@Description	Delete a server by its ID
//	@Tags			Server
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Server ID"
//	@Success		200	{object}	httpresponse.Response		"Server deleted successfully"
//	@Failure		400	{object}	httpresponse.Response		"Server ID is required"
//	@Failure		401	{object}	httpresponse.Response		"Unauthorized"
//	@Failure		403	{object}	httpresponse.Response		"Forbidden"
//	@Failure		500	{object}	httpresponse.Response		"Failed to delete server"
//	@Security		BearerAuth
//	@Router			/api/servers/{id} [delete]
func (h *serverHandler) DeleteServer(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "serverHandler.DeleteServer")
	defer span.Finish()

	serverID := c.Param("id")
	if serverID == "" {
		c.JSON(http.StatusBadRequest, httpresponse.Response{
			Message: "Server ID is required",
		})
		return
	}
	if err := h.serverUsecase.DeleteServer(ctx, serverID); err != nil {
		h.log.Errorf("Failed to delete server id=%s: %v", serverID, tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to delete server",
		})
		return
	}
	h.log.Info("Server deleted successfully ", "server_id: ", serverID)
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: "Server deleted successfully",
	})
}

// ImportServersFromExcel imports servers from Excel file
// ImportServersFromExcel godoc
//
//	@Summary		Import servers from Excel
//	@Description	Import servers from an uploaded Excel file
//	@Tags			Server
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"Excel file to import"
//	@Success		200		{object}	httpresponse.Response	"Import result with success and failure counts"
//	@Failure		400		{object}	httpresponse.Response	"File is required"
//	@Failure		401		{object}	httpresponse.Response	"Unauthorized"
//	@Failure		403		{object}	httpresponse.Response	"Forbidden"
//	@Failure		500		{object}	httpresponse.Response	"Failed to import servers"
//	@Security		BearerAuth
//	@Router			/api/servers/import [post]
func (h *serverHandler) ImportServersFromExcel(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "serverHandler.ImportServersFromExcel")
	defer span.Finish()

	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		h.log.Errorf("Failed to get file from form: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, httpresponse.Response{
			Message: "File is required",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.log.Errorf("Failed to open file: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to open file",
		})
		return
	}
	defer file.Close()

	response, err := h.serverUsecase.ImportServersFromExcel(ctx, file)
	if err != nil {
		h.log.Errorf("Failed to import servers from Excel: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to import servers",
		})
		return
	}

	h.log.Info("Servers imported successfully ",
		"filename: ", fileHeader.Filename,
		"success_count: ", response.SuccessCount,
		"failure_count: ", response.FailureCount)
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: "Servers imported successfully",
		Data:    response,
	})
}

// ExportServersToExcel exports servers to Excel file
// ExportServersToExcel godoc
//
//	@Summary		Export servers to Excel
//	@Description	Export servers to an Excel file with optional filtering
//	@Tags			Server
//	@Accept			json
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param			name	query	string	false	"Filter by server name"
//	@Param			status	query	string	false	"Filter by server status (ON/OFF)"
//	@Param			ipv4	query	string	false	"Filter by IPv4 address"
//	@Param			from	query	int		false	"Pagination offset"
//	@Param			to		query	int		false	"Pagination limit"
//	@Param			sort	query	string	false	"Sort by field (name, status, created_at, updated_at)"
//	@Param			order	query	string	false	"Sort order (asc, desc)"
//	@Success		200		{file}	file				"Excel file with servers data"
//	@Failure		400		{object}	httpresponse.Response	"Invalid request"
//	@Failure		401		{object}	httpresponse.Response	"Unauthorized"
//	@Failure		403		{object}	httpresponse.Response	"Forbidden"
//	@Failure		500		{object}	httpresponse.Response	"Failed to export servers"
//	@Security		BearerAuth
//	@Router			/api/servers/export [get]
func (h *serverHandler) ExportServersToExcel(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "serverHandler.ExportServersToExcel")
	defer span.Finish()

	var req dto.QueryServerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Errorf("Failed to bind query: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, httpresponse.Response{
			Message: "Invalid request",
		})
		return
	}

	buffer, err := h.serverUsecase.ExportServersToExcel(ctx, req.Name, req.Status, req.IPv4, req.From, req.To, req.Sort, req.Order)
	if err != nil {
		h.log.Errorf("Failed to export servers to Excel: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to export servers",
		})
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
