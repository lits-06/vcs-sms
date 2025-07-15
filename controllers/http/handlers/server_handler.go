package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/internal/domain/interfaces"
	"github.com/lits-06/vcs-sms/internal/usecases"
	"go.uber.org/zap"

	"VCS-Checkpoint1/internal/domain"
	"VCS-Checkpoint1/internal/usecase"
)

type ServerHandler struct {
	serverUsecase usecases.ServerUsecase
	logger        interfaces.Logger
}

func NewServerHandler(serverUsecase usecase.ServerUsecase, logger interfaces.Logger) *ServerHandler {
	handlerLogger := logger.With("layer", "handlers")

	return &ServerHandler{
		serverUsecase: serverUsecase,
		logger:        handlerLogger,
	}
}

func (h *ServerHandler) CreateServer(c *gin.Context) {
	var req usecases.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: err.Error(),
		})
		return
	}

	server, err := h.serverUsecase.CreateServer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "SERVER_CREATION_FAILED",
			Message: "Failed to create server",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.ServerResponse{
		Data:    server,
		Message: "Server created successfully",
	})
}

// GetServers godoc
// @Summary Get servers with pagination, filtering, and sorting
// @Description Get a list of servers with optional filtering by status, pagination, and sorting
// @Tags servers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status" Enums(online, offline)
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param sort_order query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} domain.ServersResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Router /servers [get]
func (h *ServerHandler) GetServers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	filter := &domain.ServerFilter{
		Page:      page,
		Limit:     limit,
		Status:    status,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	servers, total, err := h.serverUsecase.GetServers(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get servers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "SERVERS_FETCH_FAILED",
			Message: "Failed to fetch servers",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.ServersResponse{
		Data: servers,
		Meta: domain.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: (total + limit - 1) / limit,
		},
		Message: "Servers fetched successfully",
	})
}

// GetServer godoc
// @Summary Get a server by ID
// @Description Get detailed information about a specific server
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Success 200 {object} domain.ServerResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Router /servers/{id} [get]
func (h *ServerHandler) GetServer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid server ID",
			Details: err.Error(),
		})
		return
	}

	server, err := h.serverUsecase.GetServerByID(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrServerNotFound {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    "SERVER_NOT_FOUND",
				Message: "Server not found",
				Details: err.Error(),
			})
			return
		}
		h.logger.Error("Failed to get server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "SERVER_FETCH_FAILED",
			Message: "Failed to fetch server",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.ServerResponse{
		Data:    server,
		Message: "Server fetched successfully",
	})
}

// UpdateServer godoc
// @Summary Update a server
// @Description Update server information
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Param server body domain.UpdateServerRequest true "Server information"
// @Success 200 {object} domain.ServerResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Router /servers/{id} [put]
func (h *ServerHandler) UpdateServer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid server ID",
			Details: err.Error(),
		})
		return
	}

	var req domain.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: err.Error(),
		})
		return
	}

	server, err := h.serverUsecase.UpdateServer(c.Request.Context(), id, &req)
	if err != nil {
		if err == domain.ErrServerNotFound {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    "SERVER_NOT_FOUND",
				Message: "Server not found",
				Details: err.Error(),
			})
			return
		}
		h.logger.Error("Failed to update server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "SERVER_UPDATE_FAILED",
			Message: "Failed to update server",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.ServerResponse{
		Data:    server,
		Message: "Server updated successfully",
	})
}

// DeleteServer godoc
// @Summary Delete a server
// @Description Delete a server by ID
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Success 200 {object} domain.MessageResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Router /servers/{id} [delete]
func (h *ServerHandler) DeleteServer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid server ID",
			Details: err.Error(),
		})
		return
	}

	err = h.serverUsecase.DeleteServer(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrServerNotFound {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    "SERVER_NOT_FOUND",
				Message: "Server not found",
				Details: err.Error(),
			})
			return
		}
		h.logger.Error("Failed to delete server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "SERVER_DELETE_FAILED",
			Message: "Failed to delete server",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.MessageResponse{
		Message: "Server deleted successfully",
	})
}

// ImportServers godoc
// @Summary Import servers from Excel file
// @Description Import servers from an uploaded Excel file
// @Tags servers
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel file"
// @Success 200 {object} domain.ImportResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Router /servers/import [post]
func (h *ServerHandler) ImportServers(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "FILE_REQUIRED",
			Message: "Excel file is required",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()

	result, err := h.serverUsecase.ImportServers(c.Request.Context(), file, header.Filename)
	if err != nil {
		h.logger.Error("Failed to import servers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "IMPORT_FAILED",
			Message: "Failed to import servers",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.ImportResponse{
		Data:    result,
		Message: "Servers imported successfully",
	})
}

// ExportServers godoc
// @Summary Export servers to Excel file
// @Description Export all servers to an Excel file
// @Tags servers
// @Accept json
// @Produce application/octet-stream
// @Success 200 {file} file "Excel file"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Router /servers/export [get]
func (h *ServerHandler) ExportServers(c *gin.Context) {
	fileData, filename, err := h.serverUsecase.ExportServers(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to export servers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "EXPORT_FAILED",
			Message: "Failed to export servers",
			Details: err.Error(),
		})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.Writer.Write(fileData)
}
