package middleware

import (
	"net/http"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequireOrgRole(orgSvc *service.OrgService, requiredRole domain.OrgRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		authType, _ := c.Get(CtxAuthTypeKey)
		if authType == "apikey" {
			// API key bypasses user org role check
			c.Next()
			return
		}

		userIDVal, exists := c.Get(CtxUserIDKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userID := userIDVal.(uuid.UUID)
		userRoleVal, _ := c.Get(CtxUserRoleKey)
		if userRoleVal == domain.RoleSuperAdmin {
			c.Next()
			return
		}

		orgIDStr := c.Param("org_id")
		if orgIDStr == "" {
			orgIDStr = c.Query("org_id")
		}

		if orgIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "org_id parameter required"})
			c.Abort()
			return
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org_id"})
			c.Abort()
			return
		}

		if err := orgSvc.CheckOrgPermission(c.Request.Context(), orgID, userID, requiredRole); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Next()
	}
}
