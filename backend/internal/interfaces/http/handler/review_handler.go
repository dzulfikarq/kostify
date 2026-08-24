package handler

import (
	"net/http"

	"github.com/dzulfikarq/kostify/backend/internal/application"
	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	uc        *application.ReviewUsecase
	wishlist  *application.WishlistUsecase
	dashboard *application.DashboardUsecase
}

func NewReviewHandler(uc *application.ReviewUsecase, w *application.WishlistUsecase, d *application.DashboardUsecase) *ReviewHandler {
	return &ReviewHandler{uc: uc, wishlist: w, dashboard: d}
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required"`
	Comment string `json:"comment"`
}

func (h *ReviewHandler) Create(c *gin.Context) {
	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	rev, err := h.uc.Create(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		c.Param("id"),
		application.ReviewInput{Rating: req.Rating, Comment: req.Comment},
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusCreated, rev, "Review berhasil dikirim")
}

type ReviewItem struct {
	ID         string  `json:"id"`
	Rating     int     `json:"rating"`
	Comment    *string `json:"comment"`
	TenantName string  `json:"tenant_name"`
	CreatedAt  string  `json:"created_at"`
}

func (h *ReviewHandler) ListByProperty(c *gin.Context) {
	page, limit, err := dto.ParsePagination(c)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	f := domain.ListParams{Page: page, Limit: limit}
	f.Sort = c.Query("sort")
	f.Order = c.DefaultQuery("order", "desc")
	rows, total, err := h.uc.ListByProperty(c.Request.Context(), c.Param("id"), f)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	data := make([]ReviewItem, 0, len(rows))
	for _, r := range rows {
		data = append(data, ReviewItem{
			ID: r.ID, Rating: r.Rating, Comment: r.Comment,
			TenantName: r.TenantName, CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
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

type WishlistRequest struct {
	PropertyID string `json:"property_id" binding:"required,uuid"`
}

func (h *ReviewHandler) WishlistAdd(c *gin.Context) {
	var req WishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	if err := h.wishlist.Add(c.Request.Context(), c.GetString(middleware.ContextUserID), req.PropertyID); err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, gin.H{"property_id": req.PropertyID}, "Ditambahkan ke wishlist")
}

func (h *ReviewHandler) WishlistRemove(c *gin.Context) {
	if err := h.wishlist.Remove(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("propertyId")); err != nil {
		dto.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ReviewHandler) WishlistList(c *gin.Context) {
	page, limit, err := dto.ParsePagination(c)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	rows, total, err := h.wishlist.List(c.Request.Context(), c.GetString(middleware.ContextUserID), domain.ListParams{Page: page, Limit: limit})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	data := make([]dto.PropertySummaryResponse, 0, len(rows))
	for i := range rows {
		data = append(data, dto.NewSummaryResponse(&rows[i]))
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

func (h *ReviewHandler) Dashboard(c *gin.Context) {
	out, err := h.dashboard.Get(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		domain.Role(c.GetString(middleware.ContextRole)),
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, out, "Success")
}
