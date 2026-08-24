package middleware

import (
	"crypto/subtle"

	"github.com/gin-gonic/gin"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
)

const CSRFHeader = "X-CSRF-Token"
const CSRFCookieName = "csrf_token"

func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(CSRFHeader)
		cookie, err := c.Cookie(CSRFCookieName)
		if err != nil || cookie == "" || header == "" ||
			subtle.ConstantTimeCompare([]byte(header), []byte(cookie)) != 1 {
			dto.RespondError(c, domain.Forbidden("CSRF token tidak valid"))
			return
		}
		c.Next()
	}
}
