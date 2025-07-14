package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/internal/config"
	"github.com/lits-06/vcs-sms/internal/delivery/http/handler"
	"github.com/lits-06/vcs-sms/internal/delivery/http/middleware"
	"github.com/lits-06/vcs-sms/internal/infrastructure/logger"
	"github.com/lits-06/vcs-sms/internal/usecase"
	"github.com/lits-06/vcs-sms/pkg/validator"
)

// NewRouter creates a new HTTP router
func NewRouter(usecases *usecase.UseCases, logger *logger.Logger, validator *validator.Validator, cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API handlers
	h := handler.NewHandler(usecases, logger, validator, cfg)

	// Public routes
	public := router.Group("/api/v1")
	{
		public.POST("/auth/login", h.Login)
	}

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.JWT(cfg.JWT.Secret))
	{
		// Server routes
		servers := protected.Group("/servers")
		{
			servers.POST("", middleware.RequireScope("server:create"), h.CreateServer)
			servers.GET("", middleware.RequireScope("server:read"), h.ListServers)
			servers.GET("/:id", middleware.RequireScope("server:read"), h.GetServer)
			servers.PUT("/:id", middleware.RequireScope("server:update"), h.UpdateServer)
			servers.DELETE("/:id", middleware.RequireScope("server:delete"), h.DeleteServer)
			servers.POST("/import", middleware.RequireScope("server:import"), h.ImportServers)
			servers.GET("/export", middleware.RequireScope("server:export"), h.ExportServers)
		}

		// Report routes
		reports := protected.Group("/reports")
		{
			reports.GET("/daily", middleware.RequireScope("report:read"), h.GetDailyReport)
		}
	}

	return router
}
