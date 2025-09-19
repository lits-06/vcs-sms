package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/api/middleware"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/user_service/internal/domain"
	"github.com/lits-06/vcs-sms/user_service/internal/dto"
	"github.com/opentracing/opentracing-go"
)

type userHandler struct {
	log         logger.Logger
	userUsecase domain.UseCase
	middleware  *middleware.AuthMiddleware
}

func NewUserHandler(log logger.Logger, userUsecase domain.UseCase, middleware *middleware.AuthMiddleware) *userHandler {
	return &userHandler{
		log:         log,
		userUsecase: userUsecase,
		middleware:  middleware,
	}
}

func (h *userHandler) Register(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "userHandler.Register")
	defer span.Finish()

	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := h.userUsecase.Register(ctx, &req)
	if err != nil {
		h.log.Errorf("Failed to create user: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	h.log.Info("User created successfully", "user_id", user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

func (h *userHandler) AddUserScope(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "userHandler.AddUserScope")
	defer span.Finish()

	var req dto.AddUserScopeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind query: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.userUsecase.AddUserScope(ctx, req.UserID, req.Scopes)
	if err != nil {
		h.log.Errorf("Failed to add user scope: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user scope"})
		return
	}
	h.log.Info("User scope added successfully", "user_id", req.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "User scope added successfully"})
}

func (h *userHandler) RemoveUserScope(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "userHandler.RemoveUserScope")
	defer span.Finish()

	var req dto.RemoveUserScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.userUsecase.RemoveUserScope(ctx, req.UserID, req.Scopes); err != nil {
		h.log.Errorf("Failed to remove user scope: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user scope"})
		return
	}
	h.log.Info("User scope removed successfully", "user_id", req.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "User scope removed successfully"})
}
