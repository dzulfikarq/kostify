package application

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
)

const (
	maxPhotosPerProperty = 10
	maxUploadBytes       = 5 << 20
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type PropertyUsecase struct {
	props   domain.PropertyRepository
	rooms   domain.RoomRepository
	storage domain.ObjectStorage
}

func NewPropertyUsecase(
	props domain.PropertyRepository,
	rooms domain.RoomRepository,
	storage domain.ObjectStorage,
) *PropertyUsecase {
	return &PropertyUsecase{props: props, rooms: rooms, storage: storage}
}

type PropertyInput struct {
	Name        string
	Description *string
	Address     string
	City        string
}

type PropertyDetail struct {
	Property *domain.Property
	Photos   []domain.PropertyPhoto
	Rooms    []domain.Room
}

func (uc *PropertyUsecase) Create(ctx context.Context, ownerID string, in PropertyInput) (*domain.Property, error) {
	name := strings.TrimSpace(in.Name)
	address := strings.TrimSpace(in.Address)
	city := strings.TrimSpace(in.City)
	if details := validatePropertyFields(&name, &address, &city, trimPtr(in.Description)); len(details) > 0 {
		return nil, domain.InvalidFields(details)
	}
	p := &domain.Property{
		OwnerID:     ownerID,
		Name:        strings.TrimSpace(in.Name),
		Description: trimPtr(in.Description),
		Address:     strings.TrimSpace(in.Address),
		City:        strings.TrimSpace(in.City),
		Status:      domain.PropertyDraft,
	}
	if err := uc.props.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (uc *PropertyUsecase) Update(ctx context.Context, userID, propertyID string, in PropertyUpdateInput) (*domain.Property, error) {
	p, err := uc.ownedProperty(ctx, userID, propertyID)
	if err != nil {
		return nil, err
	}
	if !domain.EditablePropertyStatus(p.Status) {
		return nil, domain.Conflict("Kost sedang menunggu verifikasi dan tidak dapat diubah")
	}
	if details := validatePropertyFields(in.Name, in.Address, in.City, in.Description); len(details) > 0 {
		return nil, domain.InvalidFields(details)
	}

	fields := map[string]any{}
	if in.Name != nil {
		fields["name"] = strings.TrimSpace(*in.Name)
	}
	if in.Address != nil {
		fields["address"] = strings.TrimSpace(*in.Address)
	}
	if in.City != nil {
		fields["city"] = strings.TrimSpace(*in.City)
	}
	if in.Description != nil {
		fields["description"] = strings.TrimSpace(*in.Description)
	}
	if len(fields) > 1 {
		if err := uc.props.Update(ctx, p.ID, fields); err != nil {
			return nil, err
		}
	}
	return uc.props.FindByID(ctx, p.ID)
}

type PropertyUpdateInput struct {
	Name        *string
	Description *string
	Address     *string
	City        *string
}

func (uc *PropertyUsecase) GetForViewer(ctx context.Context, viewerID string, role domain.Role, id string) (*PropertyDetail, error) {
	p, err := uc.props.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	canSee := role == domain.RoleSuperAdmin ||
		(role == domain.RoleOwner && p.OwnerID == viewerID) ||
		p.Status == domain.PropertyPublished
	if !canSee {
		return nil, domain.ErrNotFound
	}
	photos, err := uc.props.PhotosOf(ctx, id)
	if err != nil {
		return nil, err
	}
	rooms, err := uc.rooms.RoomsOf(ctx, id)
	if err != nil {
		return nil, err
	}
	return &PropertyDetail{Property: p, Photos: photos, Rooms: rooms}, nil
}

func (uc *PropertyUsecase) ListPublic(ctx context.Context, f domain.PropertyFilter) ([]domain.PropertyWithStats, int64, error) {
	return uc.props.ListPublic(ctx, f)
}

func (uc *PropertyUsecase) ListMine(ctx context.Context, ownerID string, f domain.ListParams) ([]domain.Property, int64, error) {
	return uc.props.ListByOwner(ctx, ownerID, f)
}

func (uc *PropertyUsecase) Delete(ctx context.Context, propertyID string) error {
	if _, err := uc.props.FindByID(ctx, propertyID); err != nil {
		return err
	}
	hasBookings, err := uc.props.HasBookings(ctx, propertyID)
	if err != nil {
		return err
	}
	if hasBookings {
		return domain.Conflict("Kost masih memiliki riwayat booking dan tidak dapat dihapus")
	}
	return uc.props.Delete(ctx, propertyID)
}

func (uc *PropertyUsecase) Submit(ctx context.Context, ownerID, propertyID string) (*domain.Property, error) {
	p, err := uc.ownedProperty(ctx, ownerID, propertyID)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.PropertyDraft && p.Status != domain.PropertyRejected {
		return nil, domain.Conflict("Hanya kost berstatus draft atau ditolak yang dapat diajukan")
	}

	var details []domain.FieldError
	if strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.Address) == "" || strings.TrimSpace(p.City) == "" {
		details = append(details, domain.FieldError{Field: "property", Message: "nama, alamat, dan kota wajib terisi"})
	}
	if p.Description == nil || strings.TrimSpace(*p.Description) == "" {
		details = append(details, domain.FieldError{Field: "description", Message: "deskripsi wajib terisi sebelum submit"})
	}
	rooms, err := uc.rooms.RoomsOf(ctx, propertyID)
	if err != nil {
		return nil, err
	}
	if len(rooms) < 1 {
		details = append(details, domain.FieldError{Field: "rooms", Message: "minimal 1 kamar harus ditambahkan"})
	}
	photos, err := uc.props.CountPhotos(ctx, propertyID)
	if err != nil {
		return nil, err
	}
	if photos < 1 {
		details = append(details, domain.FieldError{Field: "photos", Message: "minimal 1 foto harus diunggah"})
	}
	if len(details) > 0 {
		return nil, domain.InvalidFields(details)
	}

	if err := uc.props.SetStatus(ctx, propertyID, domain.PropertyPendingVerification, nil, ownerID, domain.VerifySubmitted); err != nil {
		return nil, err
	}
	return uc.props.FindByID(ctx, propertyID)
}

func (uc *PropertyUsecase) Approve(ctx context.Context, adminID, propertyID string) (*domain.Property, error) {
	p, err := uc.pendingVerification(ctx, propertyID)
	if err != nil {
		return nil, err
	}
	if err := uc.props.SetStatus(ctx, p.ID, domain.PropertyPublished, nil, adminID, domain.VerifyApproved); err != nil {
		return nil, err
	}
	return uc.props.FindByID(ctx, p.ID)
}

func (uc *PropertyUsecase) Reject(ctx context.Context, adminID, propertyID, reason string) (*domain.Property, error) {
	reason = strings.TrimSpace(reason)
	if len([]rune(reason)) < 10 {
		return nil, domain.Invalid("reason", "alasan penolakan minimal 10 karakter")
	}
	p, err := uc.pendingVerification(ctx, propertyID)
	if err != nil {
		return nil, err
	}
	if err := uc.props.SetStatus(ctx, p.ID, domain.PropertyRejected, &reason, adminID, domain.VerifyRejected); err != nil {
		return nil, err
	}
	return uc.props.FindByID(ctx, p.ID)
}

func (uc *PropertyUsecase) UploadPhoto(
	ctx context.Context,
	ownerID, propertyID string,
	src io.Reader,
	size int64,
	contentType string,
) (*domain.PropertyPhoto, error) {
	p, err := uc.ownedProperty(ctx, ownerID, propertyID)
	if err != nil {
		return nil, err
	}
	if !domain.EditablePropertyStatus(p.Status) {
		return nil, domain.Conflict("Kost sedang menunggu verifikasi; foto tidak dapat diubah")
	}
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return nil, domain.Invalid("file", "tipe file harus JPG, PNG, atau WEBP")
	}
	if size > maxUploadBytes {
		return nil, domain.Invalid("file", "ukuran file maksimal 5 MB")
	}
	count, err := uc.props.CountPhotos(ctx, propertyID)
	if err != nil {
		return nil, err
	}
	if count >= maxPhotosPerProperty {
		return nil, domain.Conflict("Maksimal 10 foto per kost")
	}

	key := fmt.Sprintf("properties/%s/%s%s", propertyID, uuid.NewString(), ext)
	if err := uc.storage.Put(ctx, key, src, size, contentType); err != nil {
		return nil, err
	}
	photo := &domain.PropertyPhoto{
		PropertyID: propertyID,
		ObjectKey:  key,
		URL:        uc.storage.PublicURL(key),
		IsPrimary:  count == 0,
		SortOrder:  int(count) + 1,
	}
	if err := uc.props.AddPhoto(ctx, photo); err != nil {
		_ = uc.storage.Delete(ctx, key)
		return nil, err
	}
	return photo, nil
}

func (uc *PropertyUsecase) RemovePhoto(ctx context.Context, ownerID, propertyID, photoID string) error {
	if _, err := uc.ownedProperty(ctx, ownerID, propertyID); err != nil {
		return err
	}
	photo, err := uc.props.FindPhoto(ctx, propertyID, photoID)
	if err != nil {
		return err
	}
	if err := uc.props.DeletePhoto(ctx, photo); err != nil {
		return err
	}
	if err := uc.storage.Delete(ctx, photo.ObjectKey); err != nil {
		return err
	}
	if photo.IsPrimary {
		return uc.props.PromotePrimaryPhoto(ctx, propertyID)
	}
	return nil
}

func (uc *PropertyUsecase) ownedProperty(ctx context.Context, userID, propertyID string) (*domain.Property, error) {
	p, err := uc.props.FindByID(ctx, propertyID)
	if err != nil {
		return nil, err
	}
	if p.OwnerID != userID {
		return nil, domain.Forbidden("Kost ini bukan milik Anda")
	}
	return p, nil
}

func (uc *PropertyUsecase) pendingVerification(ctx context.Context, propertyID string) (*domain.Property, error) {
	p, err := uc.props.FindByID(ctx, propertyID)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.PropertyPendingVerification {
		return nil, domain.Conflict("Kost tidak sedang menunggu verifikasi")
	}
	return p, nil
}

func validatePropertyFields(name, address, city, description *string) []domain.FieldError {
	var details []domain.FieldError
	if name != nil {
		if l := len([]rune(strings.TrimSpace(*name))); l < 2 || l > 150 {
			details = append(details, domain.FieldError{Field: "name", Message: "wajib 2-150 karakter"})
		}
	}
	if address != nil {
		if strings.TrimSpace(*address) == "" {
			details = append(details, domain.FieldError{Field: "address", Message: "wajib diisi"})
		} else if len([]rune(*address)) > 500 {
			details = append(details, domain.FieldError{Field: "address", Message: "maksimal 500 karakter"})
		}
	}
	if city != nil {
		if strings.TrimSpace(*city) == "" {
			details = append(details, domain.FieldError{Field: "city", Message: "wajib diisi"})
		} else if len([]rune(*city)) > 100 {
			details = append(details, domain.FieldError{Field: "city", Message: "maksimal 100 karakter"})
		}
	}
	if description != nil && len([]rune(*description)) > 3000 {
		details = append(details, domain.FieldError{Field: "description", Message: "maksimal 3000 karakter"})
	}
	return details
}

func trimPtr(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}
