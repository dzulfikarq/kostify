package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Signer struct {
	secret []byte
}

func NewSigner(secret []byte) *Signer {
	return &Signer{secret: secret}
}

func (s *Signer) Sign(userID, role string, ttl time.Duration) (token, jti string, err error) {
	id := uuid.NewString()
	now := time.Now()
	claims := &Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ID:        id,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", "", fmt.Errorf("sign token: %w", err)
	}
	return t, id, nil
}

func (s *Signer) Parse(tokenStr string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid || claims.Subject == "" {
		return nil, errors.New("token tidak valid")
	}
	return claims, nil
}
