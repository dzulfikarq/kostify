package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dzulfikarq/kostify/backend/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookingRepo struct {
	db *gorm.DB
}

func NewBookingRepo(db *gorm.DB) *BookingRepo { return &BookingRepo{db: db} }

const bookingSelect = `
 bookings.id, bookings.room_id, bookings.tenant_id, bookings.owner_id, bookings.status,
 bookings.price_per_month, bookings.lease_duration_months, bookings.start_date,
 bookings.expires_at, bookings.checked_in_at, bookings.checked_out_at,
 bookings.cancel_reason, bookings.created_at, bookings.updated_at,
 rooms.property_id, properties.name AS property_name, rooms.room_number`

func bookingQuery(db *gorm.DB) *gorm.DB {
	return db.Table("bookings").
		Select(bookingSelect).
		Joins("JOIN rooms ON rooms.id = bookings.room_id").
		Joins("JOIN properties ON properties.id = rooms.property_id")
}

func (r *BookingRepo) Create(ctx context.Context, b *domain.Booking, actorID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var room domain.Room
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&room, "id = ?", b.RoomID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}
		if room.Status != domain.RoomAvailable {
			return domain.Conflict("Kamar sedang tidak tersedia")
		}
		var prop domain.Property
		if err := tx.Select("id", "status", "owner_id").First(&prop, "id = ?", room.PropertyID).Error; err != nil {
			return err
		}
		if prop.Status != domain.PropertyPublished {
			return domain.Conflict("Kost belum dipublikasikan")
		}
		b.OwnerID = prop.OwnerID
		b.PricePerMonth = room.PricePerMonth
		b.ExpiresAt = time.Now().Add(72 * time.Hour)
		if err := tx.Create(b).Error; err != nil {
			return mapBookingErr(err)
		}
		if err := tx.Create(&domain.BookingHistory{
			BookingID: b.ID,
			ToStatus:  domain.BookingPending,
			ChangedBy: &actorID,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.Room{}).Where("id = ?", room.ID).
			Update("status", domain.RoomPending).Error; err != nil {
			return err
		}
		title := "Booking baru masuk"
		body := fmt.Sprintf("Tenant mengajukan booking kamar %s. Respons maksimal 3 hari.", room.RoomNumber)
		ref := fmt.Sprintf(`{"booking_id":%q}`, b.ID)
		return tx.Create(&domain.Notification{
			UserID:  b.OwnerID,
			Title:   title,
			Body:    body,
			Type:    domain.NotifBooking,
			RefData: &ref,
		}).Error
	})
}

func mapBookingErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_one_pending_per_room" {
		return domain.Conflict("Anda masih memiliki booking pending untuk kamar ini")
	}
	return err
}

func (r *BookingRepo) FindByID(ctx context.Context, id string) (*domain.BookingWithRefs, error) {
	var row domain.BookingWithRefs
	err := bookingQuery(r.db.WithContext(ctx)).First(&row, "bookings.id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return &row, err
}

func (r *BookingRepo) list(ctx context.Context, where string, args []any, f domain.ListParams, propertyID string) ([]domain.BookingWithRefs, int64, error) {
	q := bookingQuery(r.db.WithContext(ctx)).Where(where, args...)
	if propertyID != "" {
		q = q.Where("rooms.property_id = ?", propertyID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "bookings.created_at DESC"
	switch f.Sort {
	case "expires_at":
		order = "bookings.expires_at ASC"
	case "created_at":
		if f.Order == "asc" {
			order = "bookings.created_at ASC"
		}
	}
	var rows []domain.BookingWithRefs
	err := q.Order(order).
		Limit(f.Limit).Offset((f.Page - 1) * f.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func (r *BookingRepo) ListByTenant(ctx context.Context, tenantID string, f domain.ListParams) ([]domain.BookingWithRefs, int64, error) {
	return r.list(ctx, "bookings.tenant_id = ?", []any{tenantID}, f, "")
}

func (r *BookingRepo) ListByOwner(ctx context.Context, ownerID string, f domain.ListParams, propertyID string) ([]domain.BookingWithRefs, int64, error) {
	return r.list(ctx, "bookings.owner_id = ?", []any{ownerID}, f, propertyID)
}

func (r *BookingRepo) Transition(ctx context.Context, t domain.BookingTransition) (*domain.Booking, error) {
	var out *domain.Booking
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var b domain.Booking
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&b, "id = ?", t.BookingID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}
		if b.Status != t.From {
			return domain.Conflict(fmt.Sprintf(
				"Aksi hanya berlaku untuk booking berstatus %s (status saat ini: %s)", t.From, b.Status))
		}

		updates := map[string]any{
			"status":     t.To,
			"updated_at": time.Now(),
		}
		if t.StartDate != nil {
			updates["start_date"] = *t.StartDate
		}
		if t.SetCheckedIn {
			updates["checked_in_at"] = time.Now()
		}
		if t.SetCheckedOut {
			updates["checked_out_at"] = time.Now()
		}
		if t.CancelReason != nil {
			updates["cancel_reason"] = *t.CancelReason
		}
		res := tx.Model(&domain.Booking{}).Where("id = ? AND status = ?", b.ID, t.From).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.Conflict("Status booking telah berubah")
		}

		if t.RoomTo != nil {
			if err := tx.Model(&domain.Room{}).Where("id = ?", b.RoomID).
				Update("status", *t.RoomTo).Error; err != nil {
				return err
			}
		}

		from := string(t.From)
		hist := domain.BookingHistory{
			BookingID:  b.ID,
			FromStatus: &from,
			ToStatus:   t.To,
			Note:       t.CancelReason,
		}
		if t.ActorID != "" {
			hist.ChangedBy = &t.ActorID
		}
		if err := tx.Create(&hist).Error; err != nil {
			return err
		}

		if t.NotifyUserID != "" && t.NotifyTitle != "" {
			ref := fmt.Sprintf(`{"booking_id":%q}`, b.ID)
			if err := tx.Create(&domain.Notification{
				UserID:  t.NotifyUserID,
				Title:   t.NotifyTitle,
				Body:    t.NotifyBody,
				Type:    domain.NotifBooking,
				RefData: &ref,
			}).Error; err != nil {
				return err
			}
		}

		if err := tx.First(&b, "id = ?", b.ID).Error; err != nil {
			return err
		}
		out = &b
		return nil
	})
	return out, err
}

func (r *BookingRepo) ExpireDue(ctx context.Context, limit int) ([]domain.ExpiredBooking, error) {
	rows := []domain.ExpiredBooking{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ids := []string{}
		err := tx.Raw(`
			UPDATE bookings SET status='expired', updated_at=now()
			WHERE id IN (
				SELECT id FROM bookings
				WHERE status='pending' AND expires_at < now()
				ORDER BY expires_at ASC
				LIMIT ?
				FOR UPDATE SKIP LOCKED
			)
			RETURNING id`, limit).Scan(&ids).Error
		if err != nil {
			return err
		}
		for _, id := range ids {
			var e domain.ExpiredBooking
			if err := tx.Raw(`SELECT id, tenant_id, owner_id, room_id FROM bookings WHERE id = ?`, id).
				Scan(&e).Error; err != nil {
				return err
			}
			rows = append(rows, e)
		}
		return nil
	})
	return rows, err
}

// ponytail: FinalizeExpiry loop per baris; batch jika volume tinggi.
func (r *BookingRepo) FinalizeExpiry(ctx context.Context, e domain.ExpiredBooking) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		from := string(domain.BookingPending)
		if err := tx.Create(&domain.BookingHistory{
			BookingID:  e.ID,
			FromStatus: &from,
			ToStatus:   domain.BookingExpired,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.Room{}).Where("id = ? AND status = ?", e.RoomID, domain.RoomPending).
			Update("status", domain.RoomAvailable).Error; err != nil {
			return err
		}
		ref := fmt.Sprintf(`{"booking_id":%q}`, e.ID)
		for _, u := range []struct{ id, body string }{
			{e.TenantID, "Waktu respons pemilik sudah lewat sehingga booking Anda kedaluwarsa."},
			{e.OwnerID, "Booking tenant sudah kedaluwarsa karena tidak direspons dalam 3 hari."},
		} {
			if err := tx.Create(&domain.Notification{
				UserID:  u.id,
				Title:   "Booking kedaluwarsa",
				Body:    u.body,
				Type:    domain.NotifBooking,
				RefData: &ref,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
