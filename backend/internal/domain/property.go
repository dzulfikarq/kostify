package domain

import "time"

type PropertyStatus string

const (
	PropertyDraft               PropertyStatus = "draft"
	PropertyPendingVerification PropertyStatus = "pending_verification"
	PropertyPublished           PropertyStatus = "published"
	PropertyRejected            PropertyStatus = "rejected"
	PropertyInactive            PropertyStatus = "inactive"
)

type VerifyAction string

const (
	VerifySubmitted VerifyAction = "submitted"
	VerifyApproved  VerifyAction = "approved"
	VerifyRejected  VerifyAction = "rejected"
)

type Property struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID         string         `gorm:"column:owner_id" json:"owner_id"`
	Name            string         `gorm:"size:150" json:"name"`
	Description     *string        `json:"description"`
	Address         string         `gorm:"size:500" json:"address"`
	City            string         `gorm:"size:100" json:"city"`
	Status          PropertyStatus `json:"status"`
	RejectionReason *string        `json:"rejection_reason"`
	RatingAvg       float64        `gorm:"column:rating_avg" json:"rating_avg"`
	RatingCount     int            `gorm:"column:rating_count" json:"rating_count"`
	VerifiedBy      *string        `gorm:"column:verified_by" json:"verified_by,omitempty"`
	VerifiedAt      *time.Time     `gorm:"column:verified_at" json:"verified_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (Property) TableName() string { return "properties" }

type PropertyPhoto struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PropertyID  string    `gorm:"column:property_id" json:"property_id"`
	ObjectKey   string    `gorm:"column:object_key;size:500" json:"-"`
	URL         string    `json:"url"`
	IsPrimary   bool      `gorm:"column:is_primary" json:"is_primary"`
	SortOrder   int       `gorm:"column:sort_order" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

func (PropertyPhoto) TableName() string { return "property_photos" }

type VerificationLog struct {
	ID         string       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PropertyID string       `gorm:"column:property_id" json:"property_id"`
	ActorID    string       `gorm:"column:actor_id" json:"actor_id"`
	Action     VerifyAction `json:"action"`
	Reason     *string      `json:"reason,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
}

func (VerificationLog) TableName() string { return "verification_logs" }

func EditablePropertyStatus(s PropertyStatus) bool {
	switch s {
	case PropertyDraft, PropertyRejected, PropertyPublished, PropertyInactive:
		return true
	default:
		return false
	}
}
