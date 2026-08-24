package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type RefreshTokenStore interface {
	Issue(ctx context.Context, userID string) (token string, err error)
	Rotate(ctx context.Context, oldToken string) (userID string, newToken string, err error)
	Revoke(ctx context.Context, token string) error
}
