package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from header or generate new one
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set in context for other handlers to use
		c.Set("request_id", requestID)

		// Set in response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// LoggingMiddleware provides structured request logging
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.String("request_id", param.Keys["request_id"].(string)),
		)
		return ""
	})
}

// MockAuthMiddleware provides a simple mock authentication for development
// In production, this would be replaced with proper JWT/OAuth validation
func MockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For development/testing, allow user_id in query parameter
		if userID := c.Query("user_id"); userID != "" {
			c.Set("user_id", userID)
			c.Set("authenticated", true)
		} else {
			// Check for Bearer token (mock implementation)
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				// Mock user ID extraction from token
				// In production, this would validate and decode the JWT
				c.Set("user_id", "550e8400-e29b-41d4-a716-446655440000") // Mock user ID
				c.Set("authenticated", true)
			}
		}

		c.Next()
	}
}

// RequireAuthMiddleware ensures the request is authenticated
func RequireAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.JSON(401, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
				"details": map[string]interface{}{
					"hint": "Include 'Authorization: Bearer <token>' header or 'user_id' query parameter",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}
