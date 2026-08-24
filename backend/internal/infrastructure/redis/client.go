package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(addr string) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr: addr,
	})
}

func Ping(ctx context.Context, c *goredis.Client) error {
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return c.Ping(pingCtx).Err()
}
