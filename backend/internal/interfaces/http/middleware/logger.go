package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

const contextRequestID = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" || len(id) > 64 {
			id = newUUID()
		}
		c.Set(contextRequestID, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(contextRequestID)
		userID, _ := c.Get("userID")
		slog.Info("http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
			"request_id", requestID,
			"user_id", userID,
		)
	}
}
