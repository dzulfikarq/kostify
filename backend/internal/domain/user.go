package domain

import "time"

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleOwner      Role = "owner"
	RoleTenant     Role = "tenant"
)

func (r Role) Valid() bool {
	switch r {
	case RoleSuperAdmin, RoleOwner, RoleTenant:
		return true
	default:
		return false
	}
}

type User struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"size:100" json:"name"`
	Email        string    `gorm:"size:255;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	Role         Role      `gorm:"type:user_role" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
