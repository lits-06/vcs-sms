package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	httpresponse "github.com/lits-06/vcs-sms/pkg/http_response"
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

// Health godoc
// @Summary Health check
// @Description Check if the report service is running
// @Tags Health
// @Produce json
// @Success 200 {object} httpresponse.Response "Service is running"
// @Router /health [get]
func (h *reportHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: "Report Service is running",
	})
}

// CreateReport godoc
// @Summary Create uptime report
// @Description Create uptime report for specific date range and send via email
// @Tags Reports
// @Accept json
// @Produce json
// @Param request body dto.UptimeRequest true "Report request payload"
// @Success 200 {object} httpresponse.Response "Report created successfully"
// @Failure 400 {object} httpresponse.Response "Invalid request"
// @Failure 401 {object} httpresponse.Response "Unauthorized"
// @Failure 403 {object} httpresponse.Response "Forbidden"
// @Failure 500 {object} httpresponse.Response "Internal server error"
// @Security BearerAuth
// @Router /api/reports [post]
func (h *reportHandler) CreateReport(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "reportHandler.CreateReport")
	defer span.Finish()

	var req dto.UptimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, httpresponse.Response{
			Message: "Invalid request",
		})
		return
	}

	err := h.reportUsecase.ReportStats(ctx, req.Email, req.StartDate, req.EndDate)
	if err != nil {
		h.log.Errorf("Failed to create report: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to create report",
		})
		return
	}
	h.log.Info("Report created successfully")
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: "Report created successfully",
	})
}
