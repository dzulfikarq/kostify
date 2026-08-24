package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
)

func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "req-unknown"
	}
	hexOut := hex.EncodeToString(b)
	return hexOut[0:8] + "-" + hexOut[8:12] + "-" + hexOut[12:16] + "-" + hexOut[16:20] + "-" + hexOut[20:32]
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				dto.RespondError(c, domain.NewAPIError(500, domain.CodeInternal, "Terjadi kesalahan pada server"))
			}
		}()
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, X-Request-ID")
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
