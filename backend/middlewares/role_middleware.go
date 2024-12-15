package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole ensures the user has the correct role to access the endpoint
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role") // Assuming the role is stored in the context by the JWT middleware

		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
		c.Abort()
	}
}
