package application

import (
	"context"
	"strings"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
)

type RoomUsecase struct {
	rooms domain.RoomRepository
	props domain.PropertyRepository
}

func NewRoomUsecase(rooms domain.RoomRepository, props domain.PropertyRepository) *RoomUsecase {
	return &RoomUsecase{rooms: rooms, props: props}
}

type RoomInput struct {
	RoomNumber    string
	PricePerMonth int
	AreaM2        *int
	Description   *string
	Facilities    []string
}

type RoomUpdateInput struct {
	RoomNumber    *string
	PricePerMonth *int
	AreaM2        *int
	Description   *string
	Facilities    []string
}

func (uc *RoomUsecase) Create(ctx context.Context, ownerID, propertyID string, in RoomInput) (*domain.Room, error) {
	if _, err := uc.ownedProperty(ctx, ownerID, propertyID); err != nil {
		return nil, err
	}
	number := strings.TrimSpace(in.RoomNumber)
	if details := validateRoomFields(&number, &in.PricePerMonth, in.AreaM2, in.Description, in.Facilities); len(details) > 0 {
		return nil, domain.InvalidFields(details)
	}
	room := &domain.Room{
		PropertyID:    propertyID,
		RoomNumber:    number,
		PricePerMonth: in.PricePerMonth,
		AreaM2:        in.AreaM2,
		Description:   trimPtr(in.Description),
		Facilities:    cleanFacilities(in.Facilities),
		Status:        domain.RoomAvailable,
	}
	if err := uc.rooms.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *RoomUsecase) Update(ctx context.Context, ownerID, roomID string, in RoomUpdateInput) (*domain.Room, error) {
	room, err := uc.ownedRoom(ctx, ownerID, roomID)
	if err != nil {
		return nil, err
	}
	var details []domain.FieldError
	if in.RoomNumber != nil {
		n := strings.TrimSpace(*in.RoomNumber)
		in.RoomNumber = &n
		details = append(details, validateRoomFields(&n, nil, nil, nil, nil)...)
	} else {
		details = append(details, validateRoomFields(nil, nil, nil, nil, nil)...)
	}
	if in.PricePerMonth != nil {
		details = append(details, validateRoomFields(nil, in.PricePerMonth, nil, nil, nil)...)
	}
	if in.AreaM2 != nil && *in.AreaM2 <= 0 {
		details = append(details, domain.FieldError{Field: "area_m2", Message: "harus lebih dari 0"})
	}
	if len(details) > 0 {
		return nil, domain.InvalidFields(details)
	}

	fields := map[string]any{}
	if in.RoomNumber != nil {
		fields["room_number"] = strings.TrimSpace(*in.RoomNumber)
	}
	if in.PricePerMonth != nil {
		fields["price_per_month"] = *in.PricePerMonth
	}
	if in.AreaM2 != nil {
		fields["area_m2"] = *in.AreaM2
	}
	if in.Description != nil {
		fields["description"] = strings.TrimSpace(*in.Description)
	}
	if in.Facilities != nil {
		fields["facilities"] = cleanFacilities(in.Facilities)
	}
	if err := uc.rooms.Update(ctx, room.ID, fields); err != nil {
		return nil, err
	}
	return uc.rooms.FindByID(ctx, room.ID)
}

func (uc *RoomUsecase) Delete(ctx context.Context, ownerID, roomID string) error {
	room, err := uc.ownedRoom(ctx, ownerID, roomID)
	if err != nil {
		return err
	}
	hasBookings, err := uc.rooms.HasActiveBookings(ctx, room.ID)
	if err != nil {
		return err
	}
	if hasBookings {
		return domain.Conflict("Kamar memiliki booking aktif dan tidak dapat dihapus")
	}
	return uc.rooms.Delete(ctx, room.ID)
}

func (uc *RoomUsecase) UpdateStatus(ctx context.Context, ownerID, roomID, newStatus string) (*domain.Room, error) {
	room, err := uc.ownedRoom(ctx, ownerID, roomID)
	if err != nil {
		return nil, err
	}
	to := domain.RoomStatus(newStatus)
	switch to {
	case domain.RoomAvailable, domain.RoomMaintenance:
	default:
		return nil, domain.Invalid("status", "hanya available atau maintenance yang dapat diubah manual")
	}
	if !domain.OwnerAllowedRoomTransition(room.Status, to) {
		return nil, domain.Conflict("Perubahan status manual hanya tersedia antara available dan maintenance")
	}
	if err := uc.rooms.Update(ctx, room.ID, map[string]any{"status": to}); err != nil {
		return nil, err
	}
	return uc.rooms.FindByID(ctx, room.ID)
}

func (uc *RoomUsecase) ListByProperty(
	ctx context.Context,
	viewerID string,
	role domain.Role,
	propertyID string,
	f domain.ListParams,
	status *domain.RoomStatus,
) ([]domain.Room, int64, error) {
	p, err := uc.props.FindByID(ctx, propertyID)
	if err != nil {
		return nil, 0, err
	}
	canSee := role == domain.RoleSuperAdmin ||
		(role == domain.RoleOwner && p.OwnerID == viewerID) ||
		p.Status == domain.PropertyPublished
	if !canSee {
		return nil, 0, domain.ErrNotFound
	}
	return uc.rooms.ListByProperty(ctx, propertyID, f, status)
}

func (uc *RoomUsecase) ownedRoom(ctx context.Context, ownerID, roomID string) (*domain.Room, error) {
	room, err := uc.rooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	p, err := uc.props.FindByID(ctx, room.PropertyID)
	if err != nil {
		return nil, err
	}
	if p.OwnerID != ownerID {
		return nil, domain.Forbidden("Kamar ini bukan milik Anda")
	}
	return room, nil
}

func (uc *RoomUsecase) ownedProperty(ctx context.Context, ownerID, propertyID string) (*domain.Property, error) {
	p, err := uc.props.FindByID(ctx, propertyID)
	if err != nil {
		return nil, err
	}
	if p.OwnerID != ownerID {
		return nil, domain.Forbidden("Kost ini bukan milik Anda")
	}
	return p, nil
}

func validateRoomFields(number *string, price *int, area *int, description *string, facilities []string) []domain.FieldError {
	var details []domain.FieldError
	if number != nil {
		l := len([]rune(*number))
		if l < 1 || l > 20 {
			details = append(details, domain.FieldError{Field: "room_number", Message: "wajib 1-20 karakter"})
		}
	}
	if price != nil && *price <= 0 {
		details = append(details, domain.FieldError{Field: "price_per_month", Message: "harus lebih dari 0"})
	}
	if price != nil && *price > 50_000_000 {
		details = append(details, domain.FieldError{Field: "price_per_month", Message: "maksimal 50.000.000"})
	}
	if area != nil && *area <= 0 {
		details = append(details, domain.FieldError{Field: "area_m2", Message: "harus lebih dari 0"})
	}
	if description != nil && len([]rune(*description)) > 1000 {
		details = append(details, domain.FieldError{Field: "description", Message: "maksimal 1000 karakter"})
	}
	for _, f := range facilities {
		if !domain.ValidFacilities[strings.ToLower(strings.TrimSpace(f))] {
			details = append(details, domain.FieldError{Field: "facilities", Message: "fasilitas tidak dikenal: " + f})
			break
		}
	}
	return details
}

func cleanFacilities(in []string) domain.StringArray {
	seen := map[string]bool{}
	out := make(domain.StringArray, 0, len(in))
	for _, f := range in {
		f = strings.ToLower(strings.TrimSpace(f))
		if f != "" && domain.ValidFacilities[f] && !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	return out
}
