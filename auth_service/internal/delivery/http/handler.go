package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/auth_service/internal/domain"
	"github.com/lits-06/vcs-sms/auth_service/internal/dto"
	httpresponse "github.com/lits-06/vcs-sms/pkg/http_response"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/middleware"
	"github.com/lits-06/vcs-sms/pkg/tracing"
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

// Health godoc
// @Summary Health check
// @Description Check if the auth service is running
// @Tags Health
// @Produce json
// @Success 200 {object} httpresponse.Response "Service is running"
// @Router /health [get]
func (h *authHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: "Auth Service is running",
	})
}

// Login godoc
// @Summary Login user
// @Description Login user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} httpresponse.Response "User logged in successfully"
// @Failure 400 {object} httpresponse.Response "Invalid request"
// @Failure 500 {object} httpresponse.Response "Failed to login"
// @Router /api/auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "authHandler.Login")
	defer span.Finish()

	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, httpresponse.Response{
			Message: "Invalid request",
		})
		return
	}

	accessToken, refreshToken, err := h.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.log.Errorf("Failed to login: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to login",
		})
		return
	}
	h.log.Info("User logged in successfully", "user_email", req.Email)
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: "User logged in successfully",
		Data: map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

// RefreshAccessToken godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} httpresponse.Response "Access token refreshed successfully"
// @Failure 400 {object} httpresponse.Response "Invalid request"
// @Failure 500 {object} httpresponse.Response "Failed to refresh access token"
// @Router /api/auth/refresh [post]
func (h *authHandler) RefreshAccessToken(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "authHandler.RefreshAccessToken")
	defer span.Finish()

	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, httpresponse.Response{
			Message: "Invalid request",
		})
		return
	}

	accessToken, err := h.authUsecase.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		h.log.Errorf("Failed to refresh access token: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, httpresponse.Response{
			Message: "Failed to refresh access token",
		})
		return
	}
	c.JSON(http.StatusOK, httpresponse.Response{
		Message: "Access token refreshed successfully",
		Data: map[string]string{
			"access_token": accessToken,
		},
	})
}
