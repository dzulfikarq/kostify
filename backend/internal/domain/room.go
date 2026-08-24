package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type RoomStatus string

const (
	RoomAvailable   RoomStatus = "available"
	RoomPending     RoomStatus = "pending"
	RoomSurvey      RoomStatus = "survey"
	RoomBooked      RoomStatus = "booked"
	RoomActive      RoomStatus = "active"
	RoomMaintenance RoomStatus = "maintenance"
	RoomCompleted   RoomStatus = "completed"
)

func OwnerAllowedRoomTransition(from, to RoomStatus) bool {
	return (from == RoomAvailable && to == RoomMaintenance) ||
		(from == RoomMaintenance && to == RoomAvailable)
}

var ValidFacilities = map[string]bool{
	"ac":               true,
	"wifi":             true,
	"private_bathroom": true,
	"shared_bathroom":  true,
	"desk":             true,
	"wardrobe":         true,
	"balcony":          true,
}

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (a *StringArray) Scan(v any) error {
	switch data := v.(type) {
	case []byte:
		return json.Unmarshal(data, a)
	case string:
		return json.Unmarshal([]byte(data), a)
	case nil:
		*a = StringArray{}
		return nil
	default:
		return fmt.Errorf("tipe tidak didukung untuk StringArray: %T", v)
	}
}

type Room struct {
	ID            string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PropertyID    string      `gorm:"column:property_id" json:"property_id"`
	RoomNumber    string      `gorm:"column:room_number;size:20" json:"room_number"`
	PricePerMonth int         `gorm:"column:price_per_month" json:"price_per_month"`
	AreaM2        *int        `gorm:"column:area_m2" json:"area_m2,omitempty"`
	Description   *string     `json:"description,omitempty"`
	Facilities    StringArray `gorm:"type:jsonb" json:"facilities"`
	Status        RoomStatus  `json:"status"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func (Room) TableName() string { return "rooms" }
