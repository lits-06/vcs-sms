package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lits-06/vcs-sms/pkg/constants"
	jwtpkg "github.com/lits-06/vcs-sms/pkg/jwt"
)

type AuthMiddleware struct {
	jwtSecretKey string
}

func NewAuthMiddleware(jwtSecretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecretKey: jwtSecretKey,
	}
}

// RequireAuth validates JWT token and extracts user claims
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Parse and validate JWT token
		token, err := jwt.ParseWithClaims(tokenString, &jwtpkg.AccessClaim{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.jwtSecretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*jwtpkg.AccessClaim)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Store user info in context for later use
		c.Set("user_email", claims.Email)
		c.Set("user_scopes", claims.Scopes)
		c.Next()
	}
}

// RequireScopes checks if the user has required scopes
func (m *AuthMiddleware) RequireScopes(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated first
		userScopes, exists := c.Get("user_scopes")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		scopesList, ok := userScopes.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid scope format"})
			c.Abort()
			return
		}

		// Check if user has admin scope (admin has all permissions)
		if m.hasScope(scopesList, constants.AdminScopeAll) {
			c.Next()
			return
		}

		// Check if user has any of the required scopes
		for _, requiredScope := range scopes {
			if m.hasScope(scopesList, requiredScope) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Insufficient permissions",
			"message": fmt.Sprintf("Required scope: %s", strings.Join(scopes, " or ")),
		})
		c.Abort()
	}
}

// hasScope checks if user has a specific scope
func (m *AuthMiddleware) hasScope(userScopes []string, requiredScope string) bool {
	for _, scope := range userScopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}
