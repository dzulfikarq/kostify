package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/pkg/random"
)

const (
	keyToken  = "rt:"
	keySet    = "rtu:"
	keyMarker = "rtm:"
)

var rotateScript = goredis.NewScript(`
local v = redis.call('GET', KEYS[1])
if not v then
    local m = redis.call('GET', KEYS[2])
    if m then
        local hs = redis.call('SMEMBERS', 'rtu:' .. m)
        for _, hh in ipairs(hs) do
            redis.call('DEL', 'rt:' .. hh)
        end
        redis.call('DEL', 'rtu:' .. m, KEYS[2])
    end
    return 0
end
redis.call('DEL', KEYS[1])
redis.call('SET', KEYS[2], v, 'EX', ARGV[2])
redis.call('SREM', 'rtu:' .. v, ARGV[1])
return v
`)

type TokenStore struct {
	rdb     *goredis.Client
	ttl     time.Duration
	markTTL time.Duration
}

func NewTokenStore(rdb *goredis.Client, ttl time.Duration) *TokenStore {
	return &TokenStore{rdb: rdb, ttl: ttl, markTTL: ttl + 24*time.Hour}
}

func (s *TokenStore) Issue(ctx context.Context, userID string) (string, error) {
	raw := random.String(32)
	h := hashToken(raw)

	if err := s.rdb.Set(ctx, keyToken+h, userID, s.ttl).Err(); err != nil {
		return "", fmt.Errorf("simpan refresh token: %w", err)
	}
	setKey := keySet + userID
	if err := s.rdb.SAdd(ctx, setKey, h).Err(); err != nil {
		return "", fmt.Errorf("daftarkan sesi: %w", err)
	}
	s.rdb.Expire(ctx, setKey, s.ttl)
	return raw, nil
}

func (s *TokenStore) Rotate(ctx context.Context, oldToken string) (string, string, error) {
	h := hashToken(oldToken)
	res, err := rotateScript.Run(ctx, s.rdb,
		[]string{keyToken + h, keyMarker + h},
		h, int(s.markTTL.Seconds()),
	).Result()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return "", "", fmt.Errorf("rotasi refresh token: %w", err)
	}

	switch v := res.(type) {
	case int64:
		if v == 0 {
			return "", "", domain.ErrSessionRevoked
		}
	case string:
		newRaw, issueErr := s.Issue(ctx, v)
		if issueErr != nil {
			return "", "", issueErr
		}
		return v, newRaw, nil
	default:
		return "", "", domain.ErrSessionRevoked
	}
	return "", "", domain.ErrSessionRevoked
}

func (s *TokenStore) Revoke(ctx context.Context, token string) error {
	h := hashToken(token)
	userID, err := s.rdb.Get(ctx, keyToken+h).Result()
	if errors.Is(err, goredis.Nil) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("baca refresh token: %w", err)
	}
	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, keyToken+h)
	pipe.SRem(ctx, keySet+userID, h)
	pipe.Del(ctx, keyMarker+h)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("cabut refresh token: %w", err)
	}
	return nil
}

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}
