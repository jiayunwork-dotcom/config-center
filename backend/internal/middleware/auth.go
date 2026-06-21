package middleware

import (
	"net/http"
	"strings"

	"config-center/internal/auth"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	UserIDKey   ContextKey = "user_id"
	UsernameKey ContextKey = "username"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UsernameKey), claims.Username)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	val, exists := c.Get(string(UserIDKey))
	if !exists {
		return 0
	}
	return val.(uint)
}

func GetUsername(c *gin.Context) string {
	val, exists := c.Get(string(UsernameKey))
	if !exists {
		return ""
	}
	return val.(string)
}
