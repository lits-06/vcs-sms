// @title User Service API
// @version 1.0
// @description This is the User Service API for VCS-SMS system

// @host localhost:8001
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
	"github.com/lits-06/vcs-sms/user_service/internal/domain"
	"github.com/lits-06/vcs-sms/user_service/internal/dto"
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

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email, username and password
// @Tags User
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 200 {object} map[string]string "User registered successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Failed to register user"
// @Router /register [post]
func (h *userHandler) Register(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "userHandler.Register")
	defer span.Finish()

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := h.userUsecase.Register(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		h.log.Errorf("Failed to register user: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}
	h.log.Info("User registered successfully", "user_id", user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

// AddUserScope godoc
// @Summary Add scopes to user
// @Description Add one or more scopes to a user by email
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AddUserScopeRequest true "Add user scope request"
// @Success 200 {object} map[string]string "User scope added successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Failed to add user scope"
// @Router /user/scopes [post]
func (h *userHandler) AddUserScope(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "userHandler.AddUserScope")
	defer span.Finish()

	var req dto.AddUserScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind query: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.userUsecase.AddUserScope(ctx, req.Email, req.Scopes)
	if err != nil {
		h.log.Errorf("Failed to add user scope: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user scope"})
		return
	}
	h.log.Info("User scope added successfully", "email", req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "User scope added successfully"})
}

// RemoveUserScope godoc
// @Summary Remove scopes from user
// @Description Remove one or more scopes from a user by email
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RemoveUserScopeRequest true "Remove user scope request"
// @Success 200 {object} map[string]string "User scope removed successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Failed to remove user scope"
// @Router /user/scopes [delete]
func (h *userHandler) RemoveUserScope(c *gin.Context) {
	ctx, span := tracing.StartHttpServerTracerSpan(c, "userHandler.RemoveUserScope")
	defer span.Finish()

	var req dto.RemoveUserScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Failed to bind JSON: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.userUsecase.RemoveUserScope(ctx, req.Email, req.Scopes); err != nil {
		h.log.Errorf("Failed to remove user scope: %v", tracing.TraceWithErr(span, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user scope"})
		return
	}
	h.log.Info("User scope removed successfully", "email", req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "User scope removed successfully"})
}
