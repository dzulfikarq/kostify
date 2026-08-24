package application

import (
	"context"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
)

type ReviewUsecase struct {
	reviews  domain.ReviewRepository
	bookings domain.BookingRepository
}

func NewReviewUsecase(reviews domain.ReviewRepository, bookings domain.BookingRepository) *ReviewUsecase {
	return &ReviewUsecase{reviews: reviews, bookings: bookings}
}

type ReviewInput struct {
	Rating  int
	Comment string
}

func (uc *ReviewUsecase) Create(ctx context.Context, tenantID, bookingID string, in ReviewInput) (*domain.Review, error) {
	if in.Rating < 1 || in.Rating > 5 {
		return nil, domain.Invalid("rating", "harus antara 1-5")
	}
	b, err := uc.bookings.FindByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if b.TenantID != tenantID {
		return nil, domain.Forbidden("Booking ini bukan milik Anda")
	}
	if b.Status != domain.BookingCompleted {
		return nil, domain.Conflict("Review hanya dapat dibuat setelah masa sewa selesai")
	}
	if existing, err := uc.reviews.FindByBooking(ctx, bookingID); err == nil && existing != nil {
		return nil, domain.Conflict("Anda sudah memberikan ulasan untuk booking ini")
	}
	var comment *string
	if in.Comment != "" {
		if len([]rune(in.Comment)) > 1000 {
			return nil, domain.Invalid("comment", "maksimal 1000 karakter")
		}
		comment = &in.Comment
	}
	rev := &domain.Review{
		BookingID:  bookingID,
		TenantID:   tenantID,
		PropertyID: b.PropertyID,
		Rating:     in.Rating,
		Comment:    comment,
	}
	if err := uc.reviews.CreateWithRecompute(ctx, rev); err != nil {
		return nil, err
	}
	return rev, nil
}

func (uc *ReviewUsecase) ListByProperty(ctx context.Context, propertyID string, f domain.ListParams) ([]domain.ReviewWithTenant, int64, error) {
	return uc.reviews.ListByProperty(ctx, propertyID, f)
}
