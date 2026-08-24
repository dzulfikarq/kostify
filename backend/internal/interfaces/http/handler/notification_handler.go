package handler

import (
	"net/http"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	repo domain.NotificationRepository
}

func NewNotificationHandler(repo domain.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

func (h *NotificationHandler) List(c *gin.Context) {
	page, limit, err := dto.ParsePagination(c)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	f := domain.ListParams{Page: page, Limit: limit}
	var isRead *bool
	if v := c.Query("is_read"); v == "true" || v == "false" {
		b := v == "true"
		isRead = &b
	}
	rows, total, err := h.repo.List(c.Request.Context(), c.GetString(middleware.ContextUserID), f, isRead)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rows,
		"meta": gin.H{
			"page": page, "limit": limit, "total": total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	found, err := h.repo.MarkRead(
		c.Request.Context(),
		c.Param("id"),
		c.GetString(middleware.ContextUserID),
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	if !found {
		dto.RespondError(c, domain.ErrNotFound)
		return
	}
	dto.OK(c, http.StatusOK, nil, "Notifikasi ditandai dibaca")
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	n, err := h.repo.MarkAllRead(c.Request.Context(), c.GetString(middleware.ContextUserID))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, gin.H{"updated": n}, "Semua notifikasi ditandai dibaca")
}
