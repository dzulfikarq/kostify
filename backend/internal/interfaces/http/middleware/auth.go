package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
	"github.com/dzulfikarq/kostify/backend/internal/pkg/jwt"
)

const (
	ContextUserID = "userID"
	ContextRole   = "role"
)

func Auth(signer *jwt.Signer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			dto.RespondError(c, domain.Unauthorized("Token autentikasi tidak ditemukan"))
			return
		}
		claims, err := signer.Parse(token)
		if err != nil {
			dto.RespondError(c, domain.Unauthorized("Token tidak valid atau kedaluwarsa"))
			return
		}
		c.Set(ContextUserID, claims.Subject)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

func RequireRole(allowed ...domain.Role) gin.HandlerFunc {
	ok := make(map[domain.Role]bool, len(allowed))
	for _, r := range allowed {
		ok[r] = true
	}
	return func(c *gin.Context) {
		role := domain.Role(c.GetString(ContextRole))
		if !ok[role] {
			dto.RespondError(c, domain.Forbidden("Anda tidak memiliki akses untuk resource ini"))
			return
		}
		c.Next()
	}
}
