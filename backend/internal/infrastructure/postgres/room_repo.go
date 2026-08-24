package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	gormio "gorm.io/gorm"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
)

type RoomRepo struct{ db *gormio.DB }

func NewRoomRepo(db *gormio.DB) *RoomRepo {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) Create(ctx context.Context, room *domain.Room) error {
	if err := r.db.WithContext(ctx).Create(room).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Conflict("Nomor kamar sudah digunakan di kost ini")
		}
		return err
	}
	return nil
}

func (r *RoomRepo) Update(ctx context.Context, id string, fields map[string]any) error {
	fields["updated_at"] = gormio.Expr("now()")
	res := r.db.WithContext(ctx).Model(&domain.Room{}).Where("id = ?", id).Updates(fields)
	if res.Error == nil && res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return res.Error
}

func (r *RoomRepo) FindByID(ctx context.Context, id string) (*domain.Room, error) {
	var room domain.Room
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&room).Error; err != nil {
		if errors.Is(err, gormio.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepo) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&domain.Room{}, "id = ?", id)
	if res.Error == nil && res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return res.Error
}

func (r *RoomRepo) RoomsOf(ctx context.Context, propertyID string) ([]domain.Room, error) {
	rooms := []domain.Room{}
	err := r.db.WithContext(ctx).
		Where("property_id = ?", propertyID).
		Order("room_number ASC").
		Find(&rooms).Error
	return rooms, err
}

func (r *RoomRepo) ListByProperty(
	ctx context.Context,
	propertyID string,
	f domain.ListParams,
	status *domain.RoomStatus,
) ([]domain.Room, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Room{}).Where("property_id = ?", propertyID)
	if f.Search != "" {
		q = q.Where("room_number ILIKE ?", "%"+f.Search+"%")
	}
	if status != nil {
		q = q.Where("status = ?", string(*status))
	}
	var total int64
	if err := q.Session(&gormio.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := []domain.Room{}
	err := q.Order("room_number ASC").
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *RoomRepo) HasActiveBookings(ctx context.Context, roomID string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(
		`SELECT EXISTS(SELECT 1 FROM bookings WHERE room_id = ? AND status IN ('pending','survey','booked','active'))`,
		roomID,
	).Scan(&exists).Error
	return exists, err
}
