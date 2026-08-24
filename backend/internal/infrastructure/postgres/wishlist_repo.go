package postgres

import (
	"context"

	"github.com/dzulfikarq/kostify/backend/internal/domain"

	"gorm.io/gorm"
)

type WishlistRepo struct {
	db *gorm.DB
}

func NewWishlistRepo(db *gorm.DB) *WishlistRepo { return &WishlistRepo{db: db} }

// ponytail: ON CONFLICT DO NOTHING → idempotent, tanpa SELECT dulu.
func (r *WishlistRepo) Add(ctx context.Context, userID, propertyID string) error {
	return r.db.WithContext(ctx).Exec(
		`INSERT INTO wishlists (user_id, property_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		userID, propertyID,
	).Error
}

func (r *WishlistRepo) Remove(ctx context.Context, userID, propertyID string) error {
	return r.db.WithContext(ctx).Exec(
		`DELETE FROM wishlists WHERE user_id = ? AND property_id = ?`,
		userID, propertyID,
	).Error
}

func (r *WishlistRepo) List(ctx context.Context, userID string, f domain.ListParams) ([]domain.PropertyWithStats, int64, error) {
	q := r.db.WithContext(ctx).
		Table("wishlists").
		Joins("JOIN properties ON properties.id = wishlists.property_id").
		Where("wishlists.user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []domain.PropertyWithStats
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT p.id, p.owner_id, p.name, p.description, p.address, p.city, p.status,
			       p.rejection_reason, p.rating_avg, p.rating_count,
			       p.created_at, p.updated_at,
			       (SELECT MIN(price_per_month)::int FROM rooms r WHERE r.property_id = p.id ) AS starting_price,
			       (SELECT COUNT(*)::int FROM rooms r WHERE r.property_id = p.id AND r.status = 'available') AS available_rooms,
			       (SELECT pp.url FROM property_photos pp WHERE pp.property_id = p.id ORDER BY is_primary DESC, sort_order ASC LIMIT 1) AS photo_url
			FROM wishlists w
			JOIN properties p ON p.id = w.property_id
			WHERE w.user_id = ?
			ORDER BY w.created_at DESC
			LIMIT ? OFFSET ?`,
			userID, f.Limit, (f.Page-1)*f.Limit,
		).Scan(&rows).Error
	return rows, total, err
}

func (r *WishlistRepo) CountByUser(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("wishlists").Where("user_id = ?", userID).Count(&n).Error
	return n, err
}
