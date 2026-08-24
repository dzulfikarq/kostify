package handler

import (
	"net/http"

	"github.com/dzulfikarq/kostify/backend/internal/application"
	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	uc *application.RoomUsecase
}

func NewRoomHandler(uc *application.RoomUsecase) *RoomHandler {
	return &RoomHandler{uc: uc}
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	room, err := h.uc.Create(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("id"), application.RoomInput{
		RoomNumber:    req.RoomNumber,
		PricePerMonth: req.PricePerMonth,
		AreaM2:        req.AreaM2,
		Description:   req.Description,
		Facilities:    req.Facilities,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusCreated, dto.NewRoomResponse(room), "Kamar berhasil ditambahkan")
}

func (h *RoomHandler) List(c *gin.Context) {
	page, limit, err := dto.ParsePagination(c)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	params := domain.ListParams{Page: page, Limit: limit, Search: c.Query("search")}
	var status *domain.RoomStatus
	if s := c.Query("status"); s != "" {
		st := domain.RoomStatus(s)
		switch st {
		case domain.RoomAvailable, domain.RoomPending, domain.RoomSurvey, domain.RoomBooked,
			domain.RoomActive, domain.RoomMaintenance, domain.RoomCompleted:
			status = &st
		default:
			dto.RespondError(c, domain.Invalid("status", "status kamar tidak dikenal"))
			return
		}
	}
	rows, total, err := h.uc.ListByProperty(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		domain.Role(c.GetString(middleware.ContextRole)),
		c.Param("id"),
		params,
		status,
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	data := make([]dto.RoomResponse, 0, len(rows))
	for i := range rows {
		data = append(data, dto.NewRoomResponse(&rows[i]))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *RoomHandler) Update(c *gin.Context) {
	var req dto.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	room, err := h.uc.Update(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("id"), application.RoomUpdateInput{
		RoomNumber:    req.RoomNumber,
		PricePerMonth: req.PricePerMonth,
		AreaM2:        req.AreaM2,
		Description:   req.Description,
		Facilities:    req.Facilities,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewRoomResponse(room), "Kamar berhasil diperbarui")
}

func (h *RoomHandler) Delete(c *gin.Context) {
	if err := h.uc.Delete(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("id")); err != nil {
		dto.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *RoomHandler) UpdateStatus(c *gin.Context) {
	var req dto.RoomStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	room, err := h.uc.UpdateStatus(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		c.Param("id"),
		req.Status,
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewRoomResponse(room), "Status kamar diperbarui")
}
