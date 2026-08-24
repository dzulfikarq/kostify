package dto

import "github.com/dzulfikarq/kostify/backend/internal/domain"

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,max=100"`
	Email    string `json:"email" binding:"required,max=255"`
	Password string `json:"password" binding:"required,max=72"`
	Role     string `json:"role" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,max=255"`
	Password string `json:"password" binding:"required,max=72"`
}

type UserResponse struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Role      domain.Role `json:"role"`
	CreatedAt string      `json:"created_at"`
}

func NewUserResponse(u *domain.User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type AuthData struct {
	User         *UserResponse `json:"user"`
	CSRFToken    string        `json:"csrf_token"`
}
