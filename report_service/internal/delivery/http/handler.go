package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/api/middleware"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/report_service/internal/domain"
	"github.com/opentracing/opentracing-go"
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

func (h *reportHandler) CreateReport(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "reportHandler.CreateReport")
	defer span.Finish()

	var req domain.UptimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.reportUsecase.ReportStats(ctx, &req)
	if err != nil {
		h.log.Errorf("Failed to create report: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
		return
	}
	h.log.Info("Report created successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Report created successfully"})
}
