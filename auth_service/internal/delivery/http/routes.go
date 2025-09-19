package http

import "github.com/gin-gonic/gin"

func (h *authHandler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/api/auth")
	auth.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Auth Service is running",
		})
	})
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.RefreshAccessToken)
}
