package postgres

import (
	"context"

	"github.com/dzulfikarq/kostify/backend/internal/domain"

	"gorm.io/gorm"
)

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo { return &NotificationRepo{db: db} }

func (r *NotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *NotificationRepo) List(ctx context.Context, userID string, f domain.ListParams, isRead *bool) ([]domain.Notification, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ?", userID)
	if isRead != nil {
		q = q.Where("is_read = ?", *isRead)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []domain.Notification
	err := q.Order("created_at DESC").
		Limit(f.Limit).Offset((f.Page - 1) * f.Limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *NotificationRepo) MarkRead(ctx context.Context, id, userID string) (bool, error) {
	res := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("id = ? AND user_id = ? AND is_read = false", id, userID).
		Update("is_read", true)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true)
	return res.RowsAffected, res.Error
}
