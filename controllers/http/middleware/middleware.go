package middleware

// import (
// 	"net/http"
// 	"strings"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt/v5"
// 	"go.uber.org/zap"
// )

// type Middleware struct {
// 	logger    *zap.Logger
// 	jwtSecret string
// }

// func NewMiddleware(logger *zap.Logger, jwtSecret string) *Middleware {
// 	return &Middleware{
// 		logger:    logger,
// 		jwtSecret: jwtSecret,
// 	}
// }

// // CORS middleware
// func (m *Middleware) CORS() gin.HandlerFunc {
// 	return gin.HandlerFunc(func(c *gin.Context) {
// 		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
// 		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
// 		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
// 		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

// 		if c.Request.Method == "OPTIONS" {
// 			c.AbortWithStatus(204)
// 			return
// 		}

// 		c.Next()
// 	})
// }

// // Logger middleware
// func (m *Middleware) Logger() gin.HandlerFunc {
// 	return gin.HandlerFunc(func(c *gin.Context) {
// 		start := time.Now()
// 		path := c.Request.URL.Path
// 		raw := c.Request.URL.RawQuery

// 		c.Next()

// 		end := time.Now()
// 		latency := end.Sub(start)

// 		if raw != "" {
// 			path = path + "?" + raw
// 		}

// 		m.logger.Info("API Request",
// 			zap.String("method", c.Request.Method),
// 			zap.String("path", path),
// 			zap.Int("status", c.Writer.Status()),
// 			zap.Duration("latency", latency),
// 			zap.String("ip", c.ClientIP()),
// 			zap.String("user_agent", c.Request.UserAgent()),
// 		)
// 	})
// }

// // JWT Authentication middleware
// func (m *Middleware) JWTAuth() gin.HandlerFunc {
// 	return gin.HandlerFunc(func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
// 			c.Abort()
// 			return
// 		}

// 		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
// 		if tokenString == authHeader {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
// 			c.Abort()
// 			return
// 		}

// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			return []byte(m.jwtSecret), nil
// 		})

// 		if err != nil || !token.Valid {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 			c.Abort()
// 			return
// 		}

// 		if claims, ok := token.Claims.(jwt.MapClaims); ok {
// 			c.Set("user_id", claims["user_id"])
// 			c.Set("scopes", claims["scopes"])
// 		}

// 		c.Next()
// 	})
// }

// // Scope-based authorization middleware
// func (m *Middleware) RequireScope(requiredScope string) gin.HandlerFunc {
// 	return gin.HandlerFunc(func(c *gin.Context) {
// 		scopes, exists := c.Get("scopes")
// 		if !exists {
// 			c.JSON(http.StatusForbidden, gin.H{"error": "No scopes found"})
// 			c.Abort()
// 			return
// 		}

// 		scopesList, ok := scopes.([]interface{})
// 		if !ok {
// 			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid scopes format"})
// 			c.Abort()
// 			return
// 		}

// 		hasScope := false
// 		for _, scope := range scopesList {
// 			if scope.(string) == requiredScope {
// 				hasScope = true
// 				break
// 			}
// 		}

// 		if !hasScope {
// 			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
// 			c.Abort()
// 			return
// 		}

// 		c.Next()
// 	})
// }
