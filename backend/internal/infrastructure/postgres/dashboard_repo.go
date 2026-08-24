package postgres

import (
	"context"

	"github.com/dzulfikarq/kostify/backend/internal/domain"

	"gorm.io/gorm"
)

type DashboardRepo struct {
	db *gorm.DB
}

func NewDashboardRepo(db *gorm.DB) *DashboardRepo { return &DashboardRepo{db: db} }

// ponytail: beberapa query terpisah, bukan satu super-SQL; gabung kalau dashboard lambat.
func (r *DashboardRepo) OwnerStats(ctx context.Context, ownerID string) (map[string]any, error) {
	out := map[string]any{"role": "owner"}
	queries := []struct {
		key  string
		sql  string
		args []any
	}{
		{"total_properties", `SELECT COUNT(*) FROM properties WHERE owner_id = ?`, []any{ownerID}},
		{"published", `SELECT COUNT(*) FROM properties WHERE owner_id = ? AND status = 'published'`, []any{ownerID}},
		{"pending_verification", `SELECT COUNT(*) FROM properties WHERE owner_id = ? AND status = 'pending_verification'`, []any{ownerID}},
		{"total_rooms", `SELECT COUNT(*) FROM rooms r JOIN properties p ON p.id = r.property_id WHERE p.owner_id = ?`, []any{ownerID}},
		{"occupied_rooms", `SELECT COUNT(*) FROM rooms r JOIN properties p ON p.id = r.property_id WHERE p.owner_id = ? AND r.status = 'active'`, []any{ownerID}},
		{"bookings_pending", `SELECT COUNT(*) FROM bookings b JOIN properties p ON p.id = (SELECT property_id FROM rooms WHERE id = b.room_id) WHERE p.owner_id = ? AND b.status = 'pending'`, []any{ownerID}},
	}
	for _, q := range queries {
		var n int64
		if err := r.db.WithContext(ctx).Raw(q.sql, q.args...).Scan(&n).Error; err != nil {
			return nil, err
		}
		out[q.key] = n
	}
	var occ float64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(
			100.0 * SUM(CASE WHEN r.status = 'active' THEN 1 ELSE 0 END) / NULLIF(COUNT(*), 0), 0)
		FROM rooms r JOIN properties p ON p.id = r.property_id WHERE p.owner_id = ?`,
		ownerID).Scan(&occ).Error; err != nil {
		return nil, err
	}
	out["occupancy_rate"] = occ
	var revenue int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(b.price_per_month * b.lease_duration_months / 12.0), 0)
		FROM bookings b JOIN rooms r ON r.id = b.room_id JOIN properties p ON p.id = r.property_id
		WHERE p.owner_id = ? AND b.status = 'active'`,
		ownerID).Scan(&revenue).Error; err != nil {
		return nil, err
	}
	out["revenue_estimation_monthly"] = revenue

	var recent []domain.BookingWithRefs
	err := bookingQuery(r.db.WithContext(ctx)).
		Where("bookings.owner_id = ?", ownerID).
		Order("created_at DESC").Limit(5).Scan(&recent).Error
	if err != nil {
		return nil, err
	}
	out["recent_bookings"] = recent
	return out, nil
}

func (r *DashboardRepo) TenantStats(ctx context.Context, tenantID string) (map[string]any, error) {
	out := map[string]any{"role": "tenant"}
	var active domain.BookingWithRefs
	err := bookingQuery(r.db.WithContext(ctx)).
		Where("bookings.tenant_id = ? AND bookings.status = 'active'", tenantID).
		First(&active).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	out["active_booking"] = nil
	if err == nil {
		out["active_booking"] = active
	}
	var pending int64
	if err := r.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM bookings WHERE tenant_id = ? AND status = 'pending'`, tenantID,
	).Scan(&pending).Error; err != nil {
		return nil, err
	}
	out["pending_bookings"] = pending
	var wishlist int64
	if err := r.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM wishlists WHERE user_id = ?`, tenantID,
	).Scan(&wishlist).Error; err != nil {
		return nil, err
	}
	out["wishlist_count"] = wishlist
	var recommended []domain.PropertyWithStats
	err = r.db.WithContext(ctx).Raw(`
		SELECT p.*, 
		       (SELECT MIN(price_per_month)::int FROM rooms r WHERE r.property_id = p.id) AS starting_price,
		       (SELECT COUNT(*)::int FROM rooms r WHERE r.property_id = p.id AND r.status = 'available') AS available_rooms,
		       (SELECT pp.url FROM property_photos pp WHERE pp.property_id = p.id ORDER BY is_primary DESC, sort_order ASC LIMIT 1) AS photo_url
		FROM properties p
		WHERE p.status = 'published'
		ORDER BY p.rating_avg DESC NULLS LAST, p.rating_count DESC
		LIMIT 5`).Scan(&recommended).Error
	if err != nil {
		return nil, err
	}
	out["recommended_properties"] = recommended
	return out, nil
}

func (r *DashboardRepo) AdminStats(ctx context.Context) (map[string]any, error) {
	out := map[string]any{"role": "super_admin"}
	queries := []struct {
		key string
		sql string
	}{
		{"users_total", `SELECT COUNT(*) FROM users`},
		{"properties_total", `SELECT COUNT(*) FROM properties`},
		{"waiting_verification", `SELECT COUNT(*) FROM properties WHERE status = 'pending_verification'`},
		{"bookings_active", `SELECT COUNT(*) FROM bookings WHERE status IN ('survey','booked','active')`},
		{"bookings_this_month", `SELECT COUNT(*) FROM bookings WHERE created_at >= date_trunc('month', now())`},
	}
	for _, q := range queries {
		var n int64
		if err := r.db.WithContext(ctx).Raw(q.sql).Scan(&n).Error; err != nil {
			return nil, err
		}
		out[q.key] = n
	}
	var queue []domain.PropertyWithStats
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.*,
		       (SELECT MIN(price_per_month)::int FROM rooms r WHERE r.property_id = p.id) AS starting_price,
		       (SELECT COUNT(*)::int FROM rooms r WHERE r.property_id = p.id AND r.status = 'available') AS available_rooms,
		       (SELECT pp.url FROM property_photos pp WHERE pp.property_id = p.id ORDER BY is_primary DESC, sort_order ASC LIMIT 1) AS photo_url
		FROM properties p
		WHERE p.status = 'pending_verification'
		ORDER BY p.updated_at ASC
		LIMIT 10`).Scan(&queue).Error
	if err != nil {
		return nil, err
	}
	out["verification_queue"] = queue
	return out, nil
}
