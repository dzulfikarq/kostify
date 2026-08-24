package domain

import (
	"time"
)

type BookingStatus string

const (
	BookingPending   BookingStatus = "pending"
	BookingSurvey    BookingStatus = "survey"
	BookingBooked    BookingStatus = "booked"
	BookingActive    BookingStatus = "active"
	BookingCompleted BookingStatus = "completed"
	BookingCancelled BookingStatus = "cancelled"
	BookingRejected  BookingStatus = "rejected"
	BookingExpired   BookingStatus = "expired"
)

func (b BookingStatus) IsValid() bool {
	switch b {
	case BookingPending, BookingSurvey, BookingBooked, BookingActive,
		BookingCompleted, BookingCancelled, BookingRejected, BookingExpired:
		return true
	}
	return false
}

type Booking struct {
	ID                   string        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID               string        `gorm:"column:room_id" json:"room_id"`
	TenantID             string        `gorm:"column:tenant_id" json:"tenant_id"`
	OwnerID              string        `gorm:"column:owner_id" json:"owner_id"`
	Status               BookingStatus `json:"status"`
	PricePerMonth        int           `gorm:"column:price_per_month" json:"price_per_month"`
	LeaseDurationMonths  int           `gorm:"column:lease_duration_months" json:"lease_duration_months"`
	StartDate            *time.Time    `gorm:"column:start_date" json:"start_date,omitempty"`
	ExpiresAt            time.Time     `gorm:"column:expires_at" json:"expires_at"`
	CheckedInAt          *time.Time    `gorm:"column:checked_in_at" json:"checked_in_at,omitempty"`
	CheckedOutAt         *time.Time    `gorm:"column:checked_out_at" json:"checked_out_at,omitempty"`
	CancelReason         *string       `gorm:"column:cancel_reason" json:"cancel_reason,omitempty"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

func (Booking) TableName() string { return "bookings" }

type BookingHistory struct {
	ID         string        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BookingID  string        `gorm:"column:booking_id" json:"booking_id"`
	FromStatus *string       `gorm:"column:from_status" json:"from_status"`
	ToStatus   BookingStatus `gorm:"column:to_status" json:"to_status"`
	ChangedBy  *string       `gorm:"column:changed_by" json:"changed_by"`
	Note       *string       `json:"note,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

func (BookingHistory) TableName() string { return "booking_history" }

type NotificationType string

const (
	NotifBooking      NotificationType = "booking"
	NotifVerification NotificationType = "verification"
	NotifReview       NotificationType = "review"
	NotifSystem       NotificationType = "system"
)

type Notification struct {
	ID        string           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string           `gorm:"column:user_id" json:"user_id"`
	Title     string           `gorm:"size:200" json:"title"`
	Body      string           `json:"body"`
	Type      NotificationType `gorm:"type:notif_type" json:"type"`
	RefData   *string          `gorm:"column:ref_data;type:jsonb" json:"ref_data,omitempty"`
	IsRead    bool             `gorm:"column:is_read" json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }

type BookingRefs struct {
	PropertyName string `json:"property_name"`
	RoomNumber   string `json:"room_number"`
}

type BookingWithRefs struct {
	Booking
	PropertyID   string `gorm:"column:property_id" json:"property_id"`
	PropertyName string `gorm:"column:property_name" json:"property_name"`
	RoomNumber   string `gorm:"column:room_number" json:"room_number"`
}
