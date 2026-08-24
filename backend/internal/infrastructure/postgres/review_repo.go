package postgres

import (
	"context"
	"errors"

	"github.com/dzulfikarq/kostify/backend/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type ReviewRepo struct {
	db *gorm.DB
}

func NewReviewRepo(db *gorm.DB) *ReviewRepo { return &ReviewRepo{db: db} }

// ponytail: recompute AVG/COUNT full-scan per review; cache agregat jika listing lambat.
func (r *ReviewRepo) CreateWithRecompute(ctx context.Context, rev *domain.Review) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rev).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.Conflict("Anda sudah memberikan ulasan untuk booking ini")
			}
			return err
		}
		return tx.Exec(`
			UPDATE properties SET
				rating_avg  = COALESCE((SELECT AVG(rating::numeric(3,2)) FROM reviews WHERE property_id = ?), 0),
				rating_count = (SELECT COUNT(*) FROM reviews WHERE property_id = ?)
			WHERE id = ?`,
			rev.PropertyID, rev.PropertyID, rev.PropertyID,
		).Error
	})
}

func (r *ReviewRepo) FindByBooking(ctx context.Context, bookingID string) (*domain.Review, error) {
	var rev domain.Review
	err := r.db.WithContext(ctx).First(&rev, "booking_id = ?", bookingID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return &rev, err
}

func (r *ReviewRepo) ListByProperty(ctx context.Context, propertyID string, f domain.ListParams) ([]domain.ReviewWithTenant, int64, error) {
	q := r.db.WithContext(ctx).
		Table("reviews").
		Select("reviews.*, users.name AS tenant_name").
		Joins("JOIN users ON users.id = reviews.tenant_id").
		Where("reviews.property_id = ?", propertyID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "created_at DESC"
	if f.Sort == "rating" {
		order = "rating " + map[bool]string{true: "ASC", false: "DESC"}[f.Order == "asc"]
	}
	var rows []domain.ReviewWithTenant
	err := q.Order(order).
		Limit(f.Limit).Offset((f.Page - 1) * f.Limit).
		Scan(&rows).Error
	return rows, total, err
}
