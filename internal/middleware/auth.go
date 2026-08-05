package middleware

import (
	"net/http"
	"strings"

	"pulsewatch/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserIDKey    = "userID"
	CtxUserRoleKey  = "userRole"
	CtxProjectIDKey = "projectID"
	CtxAuthTypeKey  = "authType"
)

func AuthMiddleware(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check API Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			key, err := authSvc.ValidateApiKey(c.Request.Context(), apiKey)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
				c.Abort()
				return
			}
			c.Set(CtxProjectIDKey, key.ProjectID)
			c.Set(CtxAuthTypeKey, "apikey")
			c.Next()
			return
		}

		// 2. Check JWT Bearer Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := authSvc.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxUserRoleKey, claims.Role)
		c.Set(CtxAuthTypeKey, "jwt")
		c.Next()
	}
}
