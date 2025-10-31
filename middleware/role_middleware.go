package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware restricts access to users with specific roles
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found in token"})
			c.Abort()
			return
		}

		userRole, ok := roleValue.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role type"})
			c.Abort()
			return
		}

		// Check if the user's role matches any allowed roles
		for _, allowed := range allowedRoles {
			if strings.EqualFold(userRole, allowed) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		c.Abort()
	}
}
