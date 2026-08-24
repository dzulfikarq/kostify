package dto

import (
	"time"

	"github.com/dzulfikarq/kostify/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type CreateBookingRequest struct {
	RoomID              string `json:"room_id" binding:"required,uuid"`
	LeaseDurationMonths int    `json:"lease_duration_months" binding:"required"`
	Note                string `json:"note"`
}

type SurveyRequest struct {
	SurveyAt *time.Time `json:"survey_at"`
}

type ConfirmBookingRequest struct {
	StartDate time.Time `json:"start_date" binding:"required"`
}

type RejectBookingRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type CancelBookingRequest struct {
	Reason string `json:"reason"`
}

func NewBookingResponse(b *domain.BookingWithRefs) gin.H {
	return gin.H{
		"id":                    b.ID,
		"room_id":               b.RoomID,
		"property":              gin.H{"id": b.PropertyID, "name": b.PropertyName},
		"room_number":           b.RoomNumber,
		"tenant_id":             b.TenantID,
		"owner_id":              b.OwnerID,
		"status":                b.Status,
		"price_per_month":       b.PricePerMonth,
		"lease_duration_months": b.LeaseDurationMonths,
		"start_date":            b.StartDate,
		"expires_at":            b.ExpiresAt,
		"checked_in_at":         b.CheckedInAt,
		"checked_out_at":        b.CheckedOutAt,
		"cancel_reason":         b.CancelReason,
		"created_at":            b.CreatedAt,
	}
}
