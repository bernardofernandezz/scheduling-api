package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/bernardofernandezz/scheduling-api/internal/service"
)

// AuthMiddleware creates a middleware for authenticating requests
func AuthMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Use 'Bearer [token]'"})
			c.Abort()
			return
		}

		// Validate token
		tokenString := parts[1]
		user, err := userService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Next()
	}
}

// RoleMiddleware creates a middleware for checking user roles
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in

