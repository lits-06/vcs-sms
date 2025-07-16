package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/api/handler"
)

type Route struct {
	serverHandler *handler.ServerHandler
}

func NewRoute(serverHandler *handler.ServerHandler) *Route {
	return &Route{
		serverHandler: serverHandler,
	}
}

func (r *Route) SetupRoutes() *gin.Engine {
	router := gin.Default()

	// // Global middleware
	// router.Use(r.middleware.Logger())
	// router.Use(r.middleware.CORS())
	// router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server Management System is running",
		})
	})

	// // API v1 group
	// v1 := router.Group("/api/v1")
	// {
	// 	// Server management routes
	// 	servers := v1.Group("/servers")
	// 	servers.Use(r.middleware.JWTAuth()) // Apply JWT authentication
	// 	{
	// 		// CRUD operations
	// 		servers.POST("", r.middleware.RequireScope("server:create"), r.serverHandler.CreateServer)
	// 		servers.GET("", r.middleware.RequireScope("server:read"), r.serverHandler.ViewServer)
	// 		servers.PUT("/:id", r.middleware.RequireScope("server:update"), r.serverHandler.UpdateServer)
	// 		servers.DELETE("/:id", r.middleware.RequireScope("server:delete"), r.serverHandler.DeleteServer)

	// 		// Import/Export operations
	// 		servers.POST("/import", r.middleware.RequireScope("server:import"), r.serverHandler.ImportServersFromExcel)
	// 		servers.GET("/export", r.middleware.RequireScope("server:export"), func(c *gin.Context) {
	// 			// Export functionality - to be implemented
	// 			c.JSON(200, gin.H{"message": "Export feature coming soon"})
	// 		})
	// 	}
	// }

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Server management routes
		servers := v1.Group("/servers")
		// servers.Use(r.middleware.JWTAuth()) // Apply JWT authentication
		{
			// CRUD operations
			servers.POST("", r.serverHandler.CreateServer)
			servers.GET("", r.serverHandler.ViewServer)
			servers.PUT("/:id", r.serverHandler.UpdateServer)
			servers.DELETE("/:id", r.serverHandler.DeleteServer)

			// Import/Export operations
			servers.POST("/import", r.serverHandler.ImportServersFromExcel)
			servers.GET("/export", r.serverHandler.ExportServersToExcel)
		}
	}

	return router
}

// RegisterRoutes is a convenience method to setup and return configured router
// func RegisterRoutes(
// 	serverHandler *handler.ServerHandler,
// 	middleware *middleware.Middleware,
// 	logger logger.Logger,
// ) *gin.Engine {
// 	router := NewRouter(serverHandler, middleware, logger)
// 	return router.SetupRoutes()
// }
