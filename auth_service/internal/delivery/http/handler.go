package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/auth_service/internal/domain"
	"github.com/lits-06/vcs-sms/auth_service/internal/dto"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/middleware"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/opentracing/opentracing-go"
)

type authHandler struct {
	log         logger.Logger
	authUsecase domain.UseCase
	middleware  *middleware.AuthMiddleware
}

func NewAuthHandler(log logger.Logger, authUsecase domain.UseCase, middleware *middleware.AuthMiddleware) *authHandler {
	return &authHandler{
		log:         log,
		authUsecase: authUsecase,
		middleware:  middleware,
	}
}

func (h *authHandler) Login(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "authHandler.Login")
	defer span.Finish()

	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	accessToken, refreshToken, err := h.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.log.Errorf("Failed to login: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}
	h.log.Info("User logged in successfully", "user_email", req.Email)
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
}

func (h *authHandler) RefreshAccessToken(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "authHandler.RefreshAccessToken")
	defer span.Finish()

	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	accessToken, err := h.authUsecase.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		h.log.Errorf("Failed to refresh access token: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh access token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
