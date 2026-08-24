package domain

import "time"

type Review struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BookingID  string    `gorm:"column:booking_id" json:"booking_id"`
	TenantID   string    `gorm:"column:tenant_id" json:"tenant_id"`
	PropertyID string    `gorm:"column:property_id" json:"property_id"`
	Rating     int       `json:"rating"`
	Comment    *string   `json:"comment,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Review) TableName() string { return "reviews" }

type ReviewWithTenant struct {
	Review
	TenantName string `gorm:"column:tenant_name" json:"tenant_name"`
}
