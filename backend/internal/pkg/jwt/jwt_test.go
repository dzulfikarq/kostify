package jwt

import (
	"strings"
	"testing"
	"time"
)

func TestSignParseRoundtrip(t *testing.T) {
	s := NewSigner([]byte("test-secret-yang-panjang-sekali-32char"))
	token, jti, err := s.Sign("user-123", "tenant", time.Minute)
	if err != nil {
		t.Fatalf("sign error: %v", err)
	}
	if jti == "" {
		t.Fatal("jti tidak boleh kosong")
	}

	claims, err := s.Parse(token)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if claims.Subject != "user-123" || claims.Role != "tenant" || claims.ID != jti {
		t.Errorf("claims tidak sesuai: %+v", claims.RegisteredClaims)
	}
}

func TestParseExpired(t *testing.T) {
	s := NewSigner([]byte("test-secret-yang-panjang-sekali-32char"))
	token, _, _ := s.Sign("user-123", "owner", -time.Minute)
	if _, err := s.Parse(token); err == nil {
		t.Error("token expired harus ditolak")
	}
}

func TestParseWrongSecret(t *testing.T) {
	a := NewSigner([]byte("secret-a-panjang-minimal-32-karakter"))
	b := NewSigner([]byte("secret-b-panjang-minimal-32-karakter"))
	token, _, _ := a.Sign("u", "tenant", time.Minute)
	if _, err := b.Parse(token); err == nil {
		t.Error("token dari secret lain harus ditolak")
	}
}

func TestParseGarbage(t *testing.T) {
	s := NewSigner([]byte("test-secret-yang-panjang-sekali-32char"))
	for _, bad := range []string{"", "abc", strings.Repeat("x. y .z", 3)} {
		if _, err := s.Parse(bad); err == nil {
			t.Errorf("input %q harus ditolak", bad)
		}
	}
}
