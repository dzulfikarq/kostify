package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	gormio "gorm.io/gorm"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
)

type PropertyRepo struct{ db *gormio.DB }

func NewPropertyRepo(db *gormio.DB) *PropertyRepo {
	return &PropertyRepo{db: db}
}

func (r *PropertyRepo) Create(ctx context.Context, p *domain.Property) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PropertyRepo) Update(ctx context.Context, id string, fields map[string]any) error {
	fields["updated_at"] = gormio.Expr("now()")
	res := r.db.WithContext(ctx).Model(&domain.Property{}).Where("id = ?", id).Updates(fields)
	if res.Error == nil && res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return res.Error
}

func (r *PropertyRepo) FindByID(ctx context.Context, id string) (*domain.Property, error) {
	var p domain.Property
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		if errors.Is(err, gormio.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PropertyRepo) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&domain.Property{}, "id = ?", id)
	if res.Error == nil && res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return res.Error
}

func (r *PropertyRepo) SetStatus(
	ctx context.Context,
	id string,
	status domain.PropertyStatus,
	rejectionReason *string,
	actorID string,
	action domain.VerifyAction,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gormio.DB) error {
		fields := map[string]any{"updated_at": gormio.Expr("now()")}
		switch action {
		case domain.VerifySubmitted:
			fields["status"] = status
			fields["rejection_reason"] = nil
		case domain.VerifyApproved:
			fields["status"] = status
			fields["verified_by"] = actorID
			fields["verified_at"] = gormio.Expr("now()")
		case domain.VerifyRejected:
			fields["status"] = status
			fields["rejection_reason"] = rejectionReason
		default:
			return fmt.Errorf("aksi verifikasi tidak dikenal: %s", action)
		}
		res := tx.Model(&domain.Property{}).Where("id = ?", id).Updates(fields)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return tx.Create(&domain.VerificationLog{
			PropertyID: id,
			ActorID:    actorID,
			Action:     action,
			Reason:     rejectionReason,
		}).Error
	})
}

func (r *PropertyRepo) HasBookings(ctx context.Context, propertyID string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(
		`SELECT EXISTS(SELECT 1 FROM bookings b JOIN rooms rm ON rm.id = b.room_id WHERE rm.property_id = ?)`,
		propertyID,
	).Scan(&exists).Error
	return exists, err
}

func (r *PropertyRepo) ListPublic(ctx context.Context, f domain.PropertyFilter) ([]domain.PropertyWithStats, int64, error) {
	base := r.db.WithContext(ctx).Table("properties AS p").
		Where("p.status = ?", string(domain.PropertyPublished))

	if f.Search != "" {
		like := "%" + f.Search + "%"
		base = base.Where("(p.name ILIKE ? OR COALESCE(p.description,'') ILIKE ? OR p.address ILIKE ?)", like, like, like)
	}
	if f.City != "" {
		base = base.Where("p.city ILIKE ?", "%"+f.City+"%")
	}
	if f.MinPrice != nil {
		base = base.Where(`EXISTS (
			SELECT 1 FROM rooms rr
			WHERE rr.property_id = p.id AND rr.status = 'available' AND rr.price_per_month >= ?
		)`, *f.MinPrice)
	}
	if f.MaxPrice != nil {
		base = base.Where(`EXISTS (
			SELECT 1 FROM rooms rr
			WHERE rr.property_id = p.id AND rr.status = 'available' AND rr.price_per_month <= ?
		)`, *f.MaxPrice)
	}
	if f.MinRating != nil {
		base = base.Where("p.rating_avg >= ?", *f.MinRating)
	}
	for _, fac := range f.Facilities {
		facJSON, err := marshalFacility(fac)
		if err != nil {
			return nil, 0, err
		}
		base = base.Where(`EXISTS (
			SELECT 1 FROM rooms rr
			WHERE rr.property_id = p.id AND rr.facilities @> ?::jsonb
		)`, facJSON)
	}

	var total int64
	if err := base.Session(&gormio.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderCol, ok := map[string]string{
		"created_at": "p.created_at",
		"price":      "starting_price",
		"rating":     "p.rating_avg",
	}[f.Sort]
	if !ok {
		orderCol = "p.created_at"
	}
	orderDir := "DESC"
	if f.Order == "asc" {
		orderDir = "ASC"
	}

	rows := []domain.PropertyWithStats{}
	err := base.Select(`p.*,
		(SELECT MIN(rr.price_per_month) FROM rooms rr WHERE rr.property_id = p.id AND rr.status = 'available') AS starting_price,
		(SELECT COUNT(*) FROM rooms rr WHERE rr.property_id = p.id AND rr.status = 'available') AS available_rooms,
		(SELECT ph.url FROM property_photos ph WHERE ph.property_id = p.id ORDER BY ph.is_primary DESC, ph.sort_order ASC LIMIT 1) AS photo_url
	`).
		Order(fmt.Sprintf("%s %s NULLS LAST", orderCol, orderDir)).
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func marshalFacility(facility string) (string, error) {
	out, err := json.Marshal([]string{facility})
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (r *PropertyRepo) ListByOwner(ctx context.Context, ownerID string, f domain.ListParams) ([]domain.Property, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Property{}).Where("owner_id = ?", ownerID)
	if f.Search != "" {
		q = q.Where("name ILIKE ?", "%"+f.Search+"%")
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	var total int64
	if err := q.Session(&gormio.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := []domain.Property{}
	err := q.Order("created_at DESC").
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *PropertyRepo) CountPhotos(ctx context.Context, propertyID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.PropertyPhoto{}).
		Where("property_id = ?", propertyID).Count(&n).Error
	return n, err
}

func (r *PropertyRepo) AddPhoto(ctx context.Context, photo *domain.PropertyPhoto) error {
	return r.db.WithContext(ctx).Create(photo).Error
}

func (r *PropertyRepo) FindPhoto(ctx context.Context, propertyID, photoID string) (*domain.PropertyPhoto, error) {
	var photo domain.PropertyPhoto
	if err := r.db.WithContext(ctx).
		Where("property_id = ? AND id = ?", propertyID, photoID).
		First(&photo).Error; err != nil {
		if errors.Is(err, gormio.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &photo, nil
}

func (r *PropertyRepo) DeletePhoto(ctx context.Context, photo *domain.PropertyPhoto) error {
	res := r.db.WithContext(ctx).Delete(photo)
	if res.Error == nil && res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return res.Error
}

func (r *PropertyRepo) PromotePrimaryPhoto(ctx context.Context, propertyID string) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE property_photos SET is_primary = true
		WHERE id = (
			SELECT id FROM property_photos
			WHERE property_id = ? AND is_primary = false
			ORDER BY sort_order ASC LIMIT 1
		)
		AND NOT EXISTS (
			SELECT 1 FROM property_photos pp
			WHERE pp.property_id = ? AND pp.is_primary = true
		)`, propertyID, propertyID).Error
}

func (r *PropertyRepo) PhotosOf(ctx context.Context, propertyID string) ([]domain.PropertyPhoto, error) {
	photos := []domain.PropertyPhoto{}
	err := r.db.WithContext(ctx).
		Where("property_id = ?", propertyID).
		Order("is_primary DESC, sort_order ASC").
		Find(&photos).Error
	return photos, err
}
