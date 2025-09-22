// @title Report Service API
// @version 1.0
// @description API for generating and managing uptime reports

// @host localhost:8002
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/middleware"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/report_service/internal/domain"
	"github.com/lits-06/vcs-sms/report_service/internal/dto"
)

type reportHandler struct {
	log           logger.Logger
	reportUsecase domain.UseCase
	middleware    *middleware.AuthMiddleware
}

func NewReportHandler(log logger.Logger, reportUsecase domain.UseCase, middleware *middleware.AuthMiddleware) *reportHandler {
	return &reportHandler{
		log:           log,
		reportUsecase: reportUsecase,
		middleware:    middleware,
	}
}

// CreateReport godoc
// @Summary Create uptime report
// @Description Create uptime report for specific date range and send via email
// @Tags Reports
// @Accept json
// @Produce json
// @Param request body dto.UptimeRequest true "Report request payload"
// @Success 200 {object} map[string]interface{} "Report created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /reports [post]
func (h *reportHandler) CreateReport(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "reportHandler.CreateReport")
	defer span.Finish()

	var req dto.UptimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.reportUsecase.ReportStats(ctx, req.Email, req.StartDate, req.EndDate)
	if err != nil {
		h.log.Errorf("Failed to create report: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
		return
	}
	h.log.Info("Report created successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Report created successfully"})
}
