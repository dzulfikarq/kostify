package dto

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dzulfikarq/kostify/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type CreatePropertyRequest struct {
	Name        string  `json:"name" binding:"required,max=150"`
	Description *string `json:"description"`
	Address     string  `json:"address" binding:"required,max=500"`
	City        string  `json:"city" binding:"required,max=100"`
}

type UpdatePropertyRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=150"`
	Description *string `json:"description"`
	Address     *string `json:"address" binding:"omitempty,max=500"`
	City        *string `json:"city" binding:"omitempty,max=100"`
}

type RejectPropertyRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type CreateRoomRequest struct {
	RoomNumber    string   `json:"room_number" binding:"required,max=20"`
	PricePerMonth int      `json:"price_per_month" binding:"required"`
	AreaM2        *int     `json:"area_m2"`
	Description   *string  `json:"description"`
	Facilities    []string `json:"facilities"`
}

type UpdateRoomRequest struct {
	RoomNumber    *string  `json:"room_number" binding:"omitempty,min=1,max=20"`
	PricePerMonth *int     `json:"price_per_month"`
	AreaM2        *int     `json:"area_m2"`
	Description   *string  `json:"description"`
	Facilities    []string `json:"facilities"`
}

type RoomStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func ParsePagination(c *gin.Context) (int, int, error) {
	var details []domain.FieldError
	page := 1
	if v := c.Query("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			details = append(details, domain.FieldError{Field: "page", Message: "harus bilangan bulat >= 1"})
		} else {
			page = n
		}
	}
	limit := 12
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			details = append(details, domain.FieldError{Field: "limit", Message: "harus antara 1-100"})
		} else {
			limit = n
		}
	}
	if len(details) > 0 {
		return 0, 0, domain.InvalidFields(details)
	}
	return page, limit, nil
}

func ParseIntQuery(c *gin.Context, name string, min int, max int) (*int, error) {
	v := c.Query(name)
	if v == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < min || n > max {
		return nil, domain.Invalid(name, fmt.Sprintf("harus bilangan bulat antara %d-%d", min, max))
	}
	return &n, nil
}

func ParseFloatQuery(c *gin.Context, name string, min float64, max float64) (*float64, error) {
	v := c.Query(name)
	if v == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f < min || f > max {
		return nil, domain.Invalid(name, fmt.Sprintf("harus angka antara %.1f-%.1f", min, max))
	}
	return &f, nil
}

func ParseFacilitiesQuery(c *gin.Context) ([]string, error) {
	raw := c.Query("facilities")
	if raw == "" {
		return nil, nil
	}
	var out []string
	for _, f := range strings.Split(raw, ",") {
		f = strings.TrimSpace(strings.ToLower(f))
		if f == "" {
			continue
		}
		if !domain.ValidFacilities[f] {
			return nil, domain.Invalid("facilities", "fasilitas tidak dikenal: "+f)
		}
		out = append(out, f)
	}
	return out, nil
}

var validSorts = map[string]bool{"created_at": true, "price": true, "rating": true}

func ParseListingFilter(c *gin.Context) (domain.PropertyFilter, error) {
	var f domain.PropertyFilter
	page, limit, err := ParsePagination(c)
	if err != nil {
		return f, err
	}
	f.Page = page
	f.Limit = limit
	f.Search = c.Query("search")
	f.City = c.Query("city")

	if f.MinPrice, err = ParseIntQuery(c, "min_price", 0, 50_000_000); err != nil {
		return f, err
	}
	if f.MaxPrice, err = ParseIntQuery(c, "max_price", 0, 50_000_000); err != nil {
		return f, err
	}
	if f.MinRating, err = ParseFloatQuery(c, "min_rating", 0, 5); err != nil {
		return f, err
	}
	if f.Facilities, err = ParseFacilitiesQuery(c); err != nil {
		return f, err
	}

	f.Sort = c.Query("sort")
	if f.Sort != "" && !validSorts[f.Sort] {
		return f, domain.Invalid("sort", "hanya created_at, price, atau rating")
	}
	order := strings.ToLower(c.Query("order"))
	switch order {
	case "":
		f.Order = "desc"
	case "asc", "desc":
		f.Order = order
	default:
		return f, domain.Invalid("order", "hanya asc atau desc")
	}
	return f, nil
}

type PhotoResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
	SortOrder int    `json:"sort_order"`
}

func NewPhotoResponse(p *domain.PropertyPhoto) PhotoResponse {
	return PhotoResponse{
		ID:        p.ID,
		URL:       p.URL,
		IsPrimary: p.IsPrimary,
		SortOrder: p.SortOrder,
	}
}

type RoomResponse struct {
	ID            string             `json:"id"`
	RoomNumber    string             `json:"room_number"`
	PricePerMonth int                `json:"price_per_month"`
	AreaM2        *int               `json:"area_m2,omitempty"`
	Description   *string            `json:"description,omitempty"`
	Facilities    []string           `json:"facilities"`
	Status        domain.RoomStatus  `json:"status"`
}

func NewRoomResponse(r *domain.Room) RoomResponse {
	facilities := []string(r.Facilities)
	if facilities == nil {
		facilities = []string{}
	}
	return RoomResponse{
		ID:            r.ID,
		RoomNumber:    r.RoomNumber,
		PricePerMonth: r.PricePerMonth,
		AreaM2:        r.AreaM2,
		Description:   r.Description,
		Facilities:    facilities,
		Status:        r.Status,
	}
}

type PropertySummaryResponse struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	City           string         `json:"city"`
	Status         domain.PropertyStatus `json:"status"`
	RatingAvg      float64        `json:"rating_avg"`
	RatingCount    int            `json:"rating_count"`
	PhotoURL       *string        `json:"photo_url"`
	StartingPrice  *int           `json:"starting_price"`
	AvailableRooms int            `json:"available_rooms"`
	CreatedAt      string         `json:"created_at"`
}

func NewSummaryResponse(row *domain.PropertyWithStats) PropertySummaryResponse {
	return PropertySummaryResponse{
		ID:             row.ID,
		Name:           row.Name,
		City:           row.City,
		Status:         row.Status,
		RatingAvg:      row.RatingAvg,
		RatingCount:    row.RatingCount,
		PhotoURL:       row.PhotoURL,
		StartingPrice:  row.StartingPrice,
		AvailableRooms: row.AvailableRooms,
		CreatedAt:      row.CreatedAt.Format(timeRFC3339),
	}
}

type OwnerPropertyResponse struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	City            string                `json:"city"`
	Status          domain.PropertyStatus `json:"status"`
	RejectionReason *string               `json:"rejection_reason,omitempty"`
	RatingAvg       float64               `json:"rating_avg"`
	RatingCount     int                   `json:"rating_count"`
	CreatedAt       string                `json:"created_at"`
}

func NewOwnerPropertyResponse(p *domain.Property) OwnerPropertyResponse {
	return OwnerPropertyResponse{
		ID:              p.ID,
		Name:            p.Name,
		City:            p.City,
		Status:          p.Status,
		RejectionReason: p.RejectionReason,
		RatingAvg:       p.RatingAvg,
		RatingCount:     p.RatingCount,
		CreatedAt:       p.CreatedAt.Format(timeRFC3339),
	}
}

type ReviewsSummary struct {
	Avg   float64 `json:"avg"`
	Count int     `json:"count"`
}

type PropertyDetailResponse struct {
	ID              string                `json:"id"`
	OwnerID         string                `json:"owner_id"`
	Name            string                `json:"name"`
	Description     *string               `json:"description"`
	Address         string                `json:"address"`
	City            string                `json:"city"`
	Status          domain.PropertyStatus `json:"status"`
	RejectionReason *string               `json:"rejection_reason,omitempty"`
	ReviewsSummary  ReviewsSummary        `json:"reviews_summary"`
	Photos          []PhotoResponse       `json:"photos"`
	Rooms           []RoomResponse        `json:"rooms"`
	CreatedAt       string                `json:"created_at"`
}

func NewDetailResponse(p *domain.Property, photos []domain.PropertyPhoto, rooms []domain.Room) PropertyDetailResponse {
	photosOut := make([]PhotoResponse, 0, len(photos))
	for i := range photos {
		photosOut = append(photosOut, NewPhotoResponse(&photos[i]))
	}
	roomsOut := make([]RoomResponse, 0, len(rooms))
	for i := range rooms {
		roomsOut = append(roomsOut, NewRoomResponse(&rooms[i]))
	}
	return PropertyDetailResponse{
		ID:              p.ID,
		OwnerID:         p.OwnerID,
		Name:            p.Name,
		Description:     p.Description,
		Address:         p.Address,
		City:            p.City,
		Status:          p.Status,
		RejectionReason: p.RejectionReason,
		ReviewsSummary:  ReviewsSummary{Avg: p.RatingAvg, Count: p.RatingCount},
		Photos:          photosOut,
		Rooms:           roomsOut,
		CreatedAt:       p.CreatedAt.Format(timeRFC3339),
	}
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
