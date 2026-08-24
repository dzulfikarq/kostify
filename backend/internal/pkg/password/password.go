package password

import (
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func Hash(plain string, cost int) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(h), nil
}

func Verify(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func Weaknesses(pw string) []string {
	var w []string
	hasLetter := false
	hasDigit := false
	for _, r := range pw {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if len([]rune(pw)) < 8 {
		w = append(w, "minimal 8 karakter")
	}
	if !hasLetter {
		w = append(w, "harus mengandung huruf")
	}
	if !hasDigit {
		w = append(w, "harus mengandung angka")
	}
	return w
}
