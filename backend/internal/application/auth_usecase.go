package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/pkg/jwt"
	"github.com/dzulfikarq/kostify/backend/internal/pkg/password"
)

type AuthUsecase struct {
	users      domain.UserRepository
	tokens     domain.RefreshTokenStore
	signer     *jwt.Signer
	accessTTL  time.Duration
	bcryptCost int
}

func NewAuthUsecase(
	users domain.UserRepository,
	tokens domain.RefreshTokenStore,
	signer *jwt.Signer,
	accessTTL time.Duration,
	bcryptCost int,
) *AuthUsecase {
	return &AuthUsecase{
		users:      users,
		tokens:     tokens,
		signer:     signer,
		accessTTL:  accessTTL,
		bcryptCost: bcryptCost,
	}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Role     string
}

func (uc *AuthUsecase) Register(ctx context.Context, in RegisterInput) (*domain.User, error) {
	var details []domain.FieldError

	in.Name = strings.TrimSpace(in.Name)
	if len(in.Name) < 2 || len([]rune(in.Name)) > 100 {
		details = append(details, domain.FieldError{Field: "name", Message: "wajib 2-100 karakter"})
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if !validEmail(in.Email) {
		details = append(details, domain.FieldError{Field: "email", Message: "harus berupa email yang valid"})
	}
	for _, w := range password.Weaknesses(in.Password) {
		details = append(details, domain.FieldError{Field: "password", Message: w})
	}

	var role domain.Role
	switch domain.Role(in.Role) {
	case domain.RoleOwner:
		role = domain.RoleOwner
	case domain.RoleTenant:
		role = domain.RoleTenant
	default:
		details = append(details, domain.FieldError{Field: "role", Message: "harus owner atau tenant"})
	}
	if len(details) > 0 {
		return nil, domain.InvalidFields(details)
	}

	if _, err := uc.users.FindByEmail(ctx, in.Email); err == nil {
		return nil, domain.Conflict("Email sudah terdaftar")
	} else if !domain.IsNotFound(err) {
		return nil, err
	}

	hash, err := password.Hash(in.Password, uc.bcryptCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: hash,
		Role:         role,
	}
	if err := uc.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

type AuthResult struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
}

func (uc *AuthUsecase) Login(ctx context.Context, email, plainPassword string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		if domain.IsNotFound(err) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if !password.Verify(user.PasswordHash, plainPassword) {
		return nil, domain.ErrInvalidCredentials
	}
	return uc.issueTokens(ctx, user)
}

func (uc *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, domain.Unauthorized("Refresh token tidak ditemukan")
	}
	userID, newRefresh, err := uc.tokens.Rotate(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	user, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrSessionRevoked
	}
	return &AuthResult{User: user, AccessToken: uc.signAccessToken(user), RefreshToken: newRefresh}, nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return uc.tokens.Revoke(ctx, refreshToken)
}

func (uc *AuthUsecase) Me(ctx context.Context, userID string) (*domain.User, error) {
	return uc.users.FindByID(ctx, userID)
}

func (uc *AuthUsecase) issueTokens(ctx context.Context, user *domain.User) (*AuthResult, error) {
	refresh, err := uc.tokens.Issue(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: user, AccessToken: uc.signAccessToken(user), RefreshToken: refresh}, nil
}

func (uc *AuthUsecase) signAccessToken(user *domain.User) string {
	token, _, err := uc.signer.Sign(user.ID, string(user.Role), uc.accessTTL)
	if err != nil {
		panic(fmt.Errorf("sign access token: %w", err))
	}
	return token
}

func validEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	return at > 0 && at < len(s)-1 && !strings.ContainsAny(s, " \t")
}
