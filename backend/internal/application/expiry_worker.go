package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
)

type ExpiryWorker struct {
	repo domain.BookingRepository
}

func NewExpiryWorker(repo domain.BookingRepository) *ExpiryWorker {
	return &ExpiryWorker{repo: repo}
}

func (w *ExpiryWorker) RunOnce(ctx context.Context) int {
	due, err := w.repo.ExpireDue(ctx, 100)
	if err != nil {
		slog.Error("worker.expired_bookings query gagal", "err", err.Error())
		return 0
	}
	for _, e := range due {
		if err := w.repo.FinalizeExpiry(ctx, e); err != nil {
			slog.Error("worker.expired_bookings finalize gagal", "booking_id", e.ID, "err", err.Error())
			continue
		}
		slog.Info("worker.expired_bookings", "booking_id", e.ID)
	}
	return len(due)
}

func (w *ExpiryWorker) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.RunOnce(ctx)
		}
	}
}
