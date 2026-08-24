package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
)

type BookingUsecase struct {
	repo    domain.BookingRepository
	notifs  domain.NotificationRepository
	users   domain.UserRepository
	mailer  MailSender
}

func NewBookingUsecase(
	repo domain.BookingRepository,
	notifs domain.NotificationRepository,
	users domain.UserRepository,
	mailer MailSender,
) *BookingUsecase {
	return &BookingUsecase{repo: repo, notifs: notifs, users: users, mailer: mailer}
}

type MailSender interface {
	Send(to, subject, body string) error
}

type CreateBookingInput struct {
	RoomID              string
	LeaseDurationMonths int
	Note                string
}

func (uc *BookingUsecase) Create(ctx context.Context, tenantID string, in CreateBookingInput) (*domain.BookingWithRefs, error) {
	if in.LeaseDurationMonths < 1 || in.LeaseDurationMonths > 36 {
		return nil, domain.Invalid("lease_duration_months", "harus antara 1-36 bulan")
	}
	if len([]rune(in.Note)) > 500 {
		return nil, domain.Invalid("note", "maksimal 500 karakter")
	}
	b := &domain.Booking{
		RoomID:              in.RoomID,
		TenantID:            tenantID,
		Status:              domain.BookingPending,
		LeaseDurationMonths: in.LeaseDurationMonths,
	}
	if err := uc.repo.Create(ctx, b, tenantID); err != nil {
		return nil, err
	}
	row, err := uc.repo.FindByID(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	uc.notifyAsync(row.OwnerID, "Booking baru masuk",
		"Tenant mengajukan booking kamar "+row.RoomNumber+" pada "+row.PropertyName+".",
		"Booking Baru — Kostify", tenantID)
	return row, nil
}

func (uc *BookingUsecase) ListMine(ctx context.Context, tenantID string, f domain.ListParams) ([]domain.BookingWithRefs, int64, error) {
	if f.Status != nil && !domain.BookingStatus(*f.Status).IsValid() {
		return nil, 0, domain.Invalid("status", "status booking tidak dikenal")
	}
	return uc.repo.ListByTenant(ctx, tenantID, f)
}

func (uc *BookingUsecase) ListOwner(ctx context.Context, ownerID string, f domain.ListParams, propertyID string) ([]domain.BookingWithRefs, int64, error) {
	if f.Status != nil && !domain.BookingStatus(*f.Status).IsValid() {
		return nil, 0, domain.Invalid("status", "status booking tidak dikenal")
	}
	return uc.repo.ListByOwner(ctx, ownerID, f, propertyID)
}

func (uc *BookingUsecase) Approve(ctx context.Context, ownerID, bookingID string, surveyAt *time.Time) (*domain.BookingWithRefs, error) {
	b, err := uc.ownedBooking(ctx, ownerID, bookingID)
	if err != nil {
		return nil, err
	}
	body := "Booking Anda disetujui. Menunggu penjadwalan survei."
	if surveyAt != nil {
		body = "Booking Anda disetujui. Jadwal survei: " + surveyAt.Format("2 Jan 2006 15:04") + "."
	}
	upd, err := uc.repo.Transition(ctx, domain.BookingTransition{
		BookingID:    bookingID,
		ActorID:      ownerID,
		From:         domain.BookingPending,
		To:           domain.BookingSurvey,
		StartDate:    surveyAt,
		NotifyUserID: b.TenantID,
		NotifyTitle:  "Booking disetujui",
		NotifyBody:   body,
	})
	if err != nil {
		return nil, err
	}
	uc.notifyAsync(b.TenantID, "Booking disetujui", body, "Booking Disetujui — Kostify", ownerID)
	return uc.withRefs(ctx, upd)
}

func (uc *BookingUsecase) Reject(ctx context.Context, ownerID, bookingID, reason string) (*domain.BookingWithRefs, error) {
	if len([]rune(reason)) < 10 {
		return nil, domain.Invalid("reason", "alasan penolakan minimal 10 karakter")
	}
	b, err := uc.ownedBooking(ctx, ownerID, bookingID)
	if err != nil {
		return nil, err
	}
	from := domain.BookingPending
	if b.Status == domain.BookingSurvey {
		from = domain.BookingSurvey
	}
	upd, err := uc.repo.Transition(ctx, domain.BookingTransition{
		BookingID:    bookingID,
		ActorID:      ownerID,
		From:         from,
		To:           domain.BookingRejected,
		RoomTo:       roomTo(domain.RoomAvailable),
		CancelReason: &reason,
		NotifyUserID: b.TenantID,
		NotifyTitle:  "Booking ditolak",
		NotifyBody:   "Booking Anda ditolak pemilik. Alasan: " + reason,
	})
	if err != nil {
		return nil, err
	}
	uc.notifyAsync(b.TenantID, "Booking ditolak", "Alasan: "+reason, "Booking Ditolak — Kostify", ownerID)
	return uc.withRefs(ctx, upd)
}

func (uc *BookingUsecase) Confirm(ctx context.Context, ownerID, bookingID string, startDate time.Time) (*domain.BookingWithRefs, error) {
	b, err := uc.ownedBooking(ctx, ownerID, bookingID)
	if err != nil {
		return nil, err
	}
	upd, err := uc.repo.Transition(ctx, domain.BookingTransition{
		BookingID:    bookingID,
		ActorID:      ownerID,
		From:         domain.BookingSurvey,
		To:           domain.BookingBooked,
		StartDate:    &startDate,
		RoomTo:       roomTo(domain.RoomBooked),
		NotifyUserID: b.TenantID,
		NotifyTitle:  "Booking dikonfirmasi",
		NotifyBody:   "Kamar dibooking untuk Anda mulai " + startDate.Format("2 Jan 2006") + ". Silakan lakukan check-in.",
	})
	if err != nil {
		return nil, err
	}
	uc.notifyAsync(b.TenantID, "Booking dikonfirmasi",
		"Mulai sewa: "+startDate.Format("2 Jan 2006"),
		"Booking Dikonfirmasi — Kostify", ownerID)
	return uc.withRefs(ctx, upd)
}

func (uc *BookingUsecase) CheckIn(ctx context.Context, tenantID, bookingID string) (*domain.BookingWithRefs, error) {
	b, err := uc.ownBookingAsTenant(ctx, tenantID, bookingID)
	if err != nil {
		return nil, err
	}
	upd, err := uc.repo.Transition(ctx, domain.BookingTransition{
		BookingID:    bookingID,
		ActorID:      tenantID,
		From:         domain.BookingBooked,
		To:           domain.BookingActive,
		SetCheckedIn: true,
		RoomTo:       roomTo(domain.RoomActive),
		NotifyUserID: b.OwnerID,
		NotifyTitle:  "Tenant check-in",
		NotifyBody:   "Tenant telah melakukan check-in.",
	})
	if err != nil {
		return nil, err
	}
	uc.notifyAsync(b.OwnerID, "Tenant check-in", "Sewa kamar "+b.RoomNumber+" dimulai.", "Check-in — Kostify", tenantID)
	return uc.withRefs(ctx, upd)
}

func (uc *BookingUsecase) CheckOut(ctx context.Context, tenantID, bookingID string) (*domain.BookingWithRefs, error) {
	b, err := uc.ownBookingAsTenant(ctx, tenantID, bookingID)
	if err != nil {
		return nil, err
	}
	upd, err := uc.repo.Transition(ctx, domain.BookingTransition{
		BookingID:     bookingID,
		ActorID:       tenantID,
		From:          domain.BookingActive,
		To:            domain.BookingCompleted,
		SetCheckedOut: true,
		RoomTo:        roomTo(domain.RoomAvailable),
		NotifyUserID:  b.OwnerID,
		NotifyTitle:   "Tenant check-out",
		NotifyBody:    "Sewa kamar " + b.RoomNumber + " selesai. Kamar tersedia kembali.",
	})
	if err != nil {
		return nil, err
	}
	uc.notifyAsync(b.OwnerID, "Tenant check-out", "Sewa selesai.", "Check-out — Kostify", tenantID)
	return uc.withRefs(ctx, upd)
}

func (uc *BookingUsecase) Cancel(ctx context.Context, tenantID, bookingID, reason string) (*domain.BookingWithRefs, error) {
	b, err := uc.ownBookingAsTenant(ctx, tenantID, bookingID)
	if err != nil {
		return nil, err
	}
	from := domain.BookingPending
	if b.Status == domain.BookingSurvey {
		from = domain.BookingSurvey
	}
	upd, err := uc.repo.Transition(ctx, domain.BookingTransition{
		BookingID:    bookingID,
		ActorID:      tenantID,
		From:         from,
		To:           domain.BookingCancelled,
		RoomTo:       roomTo(domain.RoomAvailable),
		CancelReason: nonEmpty(reason),
		NotifyUserID: b.OwnerID,
		NotifyTitle:  "Booking dibatalkan tenant",
		NotifyBody:   "Tenant membatalkan booking kamar " + b.RoomNumber + ".",
	})
	if err != nil {
		return nil, err
	}
	uc.notifyAsync(b.OwnerID, "Booking dibatalkan", "Kamar "+b.RoomNumber+" tersedia kembali.", "Booking Dibatalkan — Kostify", tenantID)
	return uc.withRefs(ctx, upd)
}

func (uc *BookingUsecase) ownedBooking(ctx context.Context, ownerID, bookingID string) (*domain.BookingWithRefs, error) {
	b, err := uc.repo.FindByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if b.OwnerID != ownerID {
		return nil, domain.Forbidden("Booking ini bukan milik kost Anda")
	}
	return b, nil
}

func (uc *BookingUsecase) ownBookingAsTenant(ctx context.Context, tenantID, bookingID string) (*domain.BookingWithRefs, error) {
	b, err := uc.repo.FindByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if b.TenantID != tenantID {
		return nil, domain.Forbidden("Booking ini bukan milik Anda")
	}
	return b, nil
}

func (uc *BookingUsecase) withRefs(ctx context.Context, b *domain.Booking) (*domain.BookingWithRefs, error) {
	return uc.repo.FindByID(ctx, b.ID)
}

// ponytail: fire-and-forget email; queue asynq kalau perlu retry/durability.
func (uc *BookingUsecase) notifyAsync(userID, title, body, emailSubject, actorID string) {
	go func() {
		defer func() { _ = recover() }()
		n := &domain.Notification{UserID: userID, Title: title, Body: body, Type: domain.NotifBooking}
		if err := uc.notifs.Create(context.Background(), n); err != nil {
			slog.Error("gagal simpan notifikasi", "err", err.Error())
		}
		if uc.mailer == nil {
			return
		}
		u, err := uc.users.FindByID(context.Background(), userID)
		if err != nil || u.Email == "" {
			return
		}
		if err := uc.mailer.Send(u.Email, emailSubject, title+"\n\n"+body); err != nil {
			slog.Error("gagal kirim email", "err", err.Error())
		}
	}()
}

func roomTo(s domain.RoomStatus) *domain.RoomStatus { return &s }

func nonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
