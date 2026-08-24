package handler

import (
	"net/http"
	"time"

	"github.com/dzulfikarq/kostify/backend/internal/application"
	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	uc *application.BookingUsecase
}

func NewBookingHandler(uc *application.BookingUsecase) *BookingHandler {
	return &BookingHandler{uc: uc}
}

func (h *BookingHandler) Create(c *gin.Context) {
	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	b, err := h.uc.Create(c.Request.Context(), c.GetString(middleware.ContextUserID), application.CreateBookingInput{
		RoomID:              req.RoomID,
		LeaseDurationMonths: req.LeaseDurationMonths,
		Note:                req.Note,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusCreated, dto.NewBookingResponse(b),
		"Booking berhasil dibuat. Menunggu konfirmasi pemilik maksimal 3 hari.")
}

func parseBookingList(c *gin.Context) (domain.ListParams, error) {
	var f domain.ListParams
	page, limit, err := dto.ParsePagination(c)
	if err != nil {
		return f, err
	}
	f.Page = page
	f.Limit = limit
	if s := c.Query("status"); s != "" {
		f.Status = &s
	}
	return f, nil
}

func respondBookings(c *gin.Context, rows []domain.BookingWithRefs, total int64, page, limit int) {
	data := make([]gin.H, 0, len(rows))
	for i := range rows {
		data = append(data, dto.NewBookingResponse(&rows[i]))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"meta": gin.H{
			"page": page, "limit": limit, "total": total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *BookingHandler) ListMine(c *gin.Context) {
	f, err := parseBookingList(c)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	rows, total, err := h.uc.ListMine(c.Request.Context(), c.GetString(middleware.ContextUserID), f)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	respondBookings(c, rows, total, f.Page, f.Limit)
}

func (h *BookingHandler) ListOwner(c *gin.Context) {
	f, err := parseBookingList(c)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	rows, total, err := h.uc.ListOwner(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		f,
		c.Query("property_id"),
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	respondBookings(c, rows, total, f.Page, f.Limit)
}

func (h *BookingHandler) Approve(c *gin.Context) {
	var req dto.SurveyRequest
	_ = c.ShouldBindJSON(&req)
	b, err := h.uc.Approve(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		c.Param("id"),
		req.SurveyAt,
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewBookingResponse(b), "Booking disetujui, menunggu jadwal survey")
}

func (h *BookingHandler) Reject(c *gin.Context) {
	var req dto.RejectBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	b, err := h.uc.Reject(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		c.Param("id"),
		req.Reason,
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewBookingResponse(b), "Booking ditolak")
}

func (h *BookingHandler) Confirm(c *gin.Context) {
	var req dto.ConfirmBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.StartDate.IsZero() {
		dto.RespondError(c, domain.Invalid("start_date", "tanggal mulai sewa wajib diisi (format YYYY-MM-DD)"))
		return
	}
	sd := req.StartDate.Truncate(24 * time.Hour)
	b, err := h.uc.Confirm(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		c.Param("id"),
		sd,
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewBookingResponse(b), "Booking dikonfirmasi")
}

func (h *BookingHandler) CheckIn(c *gin.Context) {
	b, err := h.uc.CheckIn(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("id"))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewBookingResponse(b), "Check-in berhasil. Selamat menempati!")
}

func (h *BookingHandler) CheckOut(c *gin.Context) {
	b, err := h.uc.CheckOut(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("id"))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewBookingResponse(b), "Check-out berhasil. Terima kasih!")
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	var req dto.CancelBookingRequest
	_ = c.ShouldBindJSON(&req)
	b, err := h.uc.Cancel(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		c.Param("id"),
		req.Reason,
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewBookingResponse(b), "Booking dibatalkan")
}
