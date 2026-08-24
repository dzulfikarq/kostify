package domain

import (
	"context"
	"io"
)

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

type ListParams struct {
	Page   int
	Limit  int
	Search string
	Status *string
}

type PropertyFilter struct {
	ListParams
	City       string
	MinPrice   *int
	MaxPrice   *int
	MinRating  *float64
	Facilities []string
	Sort       string
	Order      string
}

type PropertyWithStats struct {
	Property
	StartingPrice  *int    `gorm:"column:starting_price"`
	AvailableRooms int     `gorm:"column:available_rooms"`
	PhotoURL       *string `gorm:"column:photo_url"`
}

type PropertyRepository interface {
	Create(ctx context.Context, p *Property) error
	Update(ctx context.Context, id string, fields map[string]any) error
	FindByID(ctx context.Context, id string) (*Property, error)
	Delete(ctx context.Context, id string) error
	SetStatus(ctx context.Context, id string, status PropertyStatus, rejectionReason *string, actorID string, action VerifyAction) error
	HasBookings(ctx context.Context, propertyID string) (bool, error)
	ListPublic(ctx context.Context, f PropertyFilter) ([]PropertyWithStats, int64, error)
	ListByOwner(ctx context.Context, ownerID string, f ListParams) ([]Property, int64, error)
	CountPhotos(ctx context.Context, propertyID string) (int64, error)
	AddPhoto(ctx context.Context, photo *PropertyPhoto) error
	FindPhoto(ctx context.Context, propertyID, photoID string) (*PropertyPhoto, error)
	DeletePhoto(ctx context.Context, photo *PropertyPhoto) error
	PromotePrimaryPhoto(ctx context.Context, propertyID string) error
	PhotosOf(ctx context.Context, propertyID string) ([]PropertyPhoto, error)
}

type RoomRepository interface {
	Create(ctx context.Context, r *Room) error
	Update(ctx context.Context, id string, fields map[string]any) error
	FindByID(ctx context.Context, id string) (*Room, error)
	Delete(ctx context.Context, id string) error
	RoomsOf(ctx context.Context, propertyID string) ([]Room, error)
	ListByProperty(ctx context.Context, propertyID string, f ListParams, status *RoomStatus) ([]Room, int64, error)
	HasActiveBookings(ctx context.Context, roomID string) (bool, error)
}

type ObjectStorage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}
