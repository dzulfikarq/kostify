package domain

import (
	"context"
	"io"
	"time"
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
	Sort   string
	Order  string
}

type PropertyFilter struct {
	ListParams
	City       string
	MinPrice   *int
	MaxPrice   *int
	MinRating  *float64
	Facilities []string
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

type BookingTransition struct {
	BookingID    string
	ActorID      string
	From         BookingStatus
	To           BookingStatus
	RoomTo       *RoomStatus
	StartDate    *time.Time
	SetCheckedIn bool
	SetCheckedOut bool
	CancelReason *string
	Note         *string

	NotifyUserID string
	NotifyTitle  string
	NotifyBody   string
	EmailTo      *string
	EmailSubject string
}

type ExpiredBooking struct {
	ID       string
	TenantID string
	OwnerID  string
	RoomID   string
}

type BookingRepository interface {
	Create(ctx context.Context, b *Booking, actorID string) error
	FindByID(ctx context.Context, id string) (*BookingWithRefs, error)
	ListByTenant(ctx context.Context, tenantID string, f ListParams) ([]BookingWithRefs, int64, error)
	ListByOwner(ctx context.Context, ownerID string, f ListParams, propertyID string) ([]BookingWithRefs, int64, error)
	Transition(ctx context.Context, t BookingTransition) (*Booking, error)
	ExpireDue(ctx context.Context, limit int) ([]ExpiredBooking, error)
	FinalizeExpiry(ctx context.Context, e ExpiredBooking) error
}

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	List(ctx context.Context, userID string, f ListParams, isRead *bool) ([]Notification, int64, error)
	MarkRead(ctx context.Context, id, userID string) (bool, error)
	MarkAllRead(ctx context.Context, userID string) (int64, error)
}

type ReviewRepository interface {
	CreateWithRecompute(ctx context.Context, r *Review) error
	FindByBooking(ctx context.Context, bookingID string) (*Review, error)
	ListByProperty(ctx context.Context, propertyID string, f ListParams) ([]ReviewWithTenant, int64, error)
}

type WishlistRepository interface {
	Add(ctx context.Context, userID, propertyID string) error
	Remove(ctx context.Context, userID, propertyID string) error
	List(ctx context.Context, userID string, f ListParams) ([]PropertyWithStats, int64, error)
	CountByUser(ctx context.Context, userID string) (int64, error)
}

type DashboardRepository interface {
	OwnerStats(ctx context.Context, ownerID string) (map[string]any, error)
	TenantStats(ctx context.Context, tenantID string) (map[string]any, error)
	AdminStats(ctx context.Context) (map[string]any, error)
}
