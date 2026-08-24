package middleware

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
)

func LoginRateLimit(rdb *goredis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rl:login:" + c.ClientIP()
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, time.Minute)
		}
		if count > 5 {
			dto.RespondError(c, domain.RateLimited("Terlalu banyak percobaan login. Coba lagi dalam 60 detik."))
			return
		}
		c.Next()
	}
}
