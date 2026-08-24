package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/dzulfikarq/kostify/backend/internal/application"
	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type PropertyHandler struct {
	uc *application.PropertyUsecase
}

func NewPropertyHandler(uc *application.PropertyUsecase) *PropertyHandler {
	return &PropertyHandler{uc: uc}
}

func (h *PropertyHandler) Create(c *gin.Context) {
	var req dto.CreatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	p, err := h.uc.Create(c.Request.Context(), c.GetString(middleware.ContextUserID), application.PropertyInput{
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		City:        req.City,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusCreated, gin.H{"id": p.ID, "status": p.Status}, "Draft kost berhasil dibuat")
}

func (h *PropertyHandler) Update(c *gin.Context) {
	var req dto.UpdatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	p, err := h.uc.Update(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("id"), application.PropertyUpdateInput{
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		City:        req.City,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, gin.H{
		"id":         p.ID,
		"name":       p.Name,
		"status":     p.Status,
		"updated_at": p.UpdatedAt,
	}, "Kost berhasil diperbarui")
}

func (h *PropertyHandler) ListPublic(c *gin.Context) {
	filter, err := dto.ParseListingFilter(c)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	rows, total, err := h.uc.ListPublic(c.Request.Context(), filter)
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
			"page":        filter.Page,
			"limit":       filter.Limit,
			"total":       total,
			"total_pages": (total + int64(filter.Limit) - 1) / int64(filter.Limit),
		},
	})
}

func (h *PropertyHandler) ListMine(c *gin.Context) {
	page, limit, err := dto.ParsePagination(c)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	params := domain.ListParams{Page: page, Limit: limit, Search: c.Query("search")}
	if s := c.Query("status"); s != "" {
		st := domain.PropertyStatus(s)
		switch st {
		case domain.PropertyDraft, domain.PropertyPendingVerification, domain.PropertyPublished,
			domain.PropertyRejected, domain.PropertyInactive:
			sv := string(st)
			params.Status = &sv
		default:
			dto.RespondError(c, domain.Invalid("status", "status tidak dikenal"))
			return
		}
	}
	rows, total, err := h.uc.ListMine(c.Request.Context(), c.GetString(middleware.ContextUserID), params)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	data := make([]dto.OwnerPropertyResponse, 0, len(rows))
	for i := range rows {
		data = append(data, dto.NewOwnerPropertyResponse(&rows[i]))
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

func (h *PropertyHandler) GetDetail(c *gin.Context) {
	viewerID, _ := c.Get(middleware.ContextUserID)
	roleStr, _ := c.Get(middleware.ContextRole)
	role, _ := roleStr.(string)

	detail, err := h.uc.GetForViewer(
		c.Request.Context(),
		viewerIDString(viewerID),
		domain.Role(role),
		c.Param("id"),
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewDetailResponse(detail.Property, detail.Photos, detail.Rooms), "Success")
}

func viewerIDString(v any) string {
	s, _ := v.(string)
	return s
}

func (h *PropertyHandler) Submit(c *gin.Context) {
	p, err := h.uc.Submit(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("id"))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, gin.H{"id": p.ID, "status": p.Status}, "Kost diajukan untuk verifikasi")
}

func (h *PropertyHandler) Approve(c *gin.Context) {
	p, err := h.uc.Approve(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("id"))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, gin.H{"id": p.ID, "status": p.Status}, "Kost disetujui dan dipublikasikan")
}

func (h *PropertyHandler) Reject(c *gin.Context) {
	var req dto.RejectPropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}
	p, err := h.uc.Reject(c.Request.Context(), c.GetString(middleware.ContextUserID), c.Param("id"), req.Reason)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, gin.H{"id": p.ID, "status": p.Status, "rejection_reason": p.RejectionReason}, "Kost ditolak")
}

func (h *PropertyHandler) AdminDelete(c *gin.Context) {
	if err := h.uc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		dto.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PropertyHandler) UploadPhoto(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		dto.RespondError(c, domain.Invalid("file", "file foto wajib diunggah"))
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		dto.RespondError(c, domain.BadRequest("File tidak dapat dibaca"))
		return
	}
	defer f.Close()

	head := make([]byte, 512)
	n, readErr := io.ReadFull(f, head)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) && !errors.Is(readErr, io.EOF) {
		dto.RespondError(c, domain.BadRequest("File tidak dapat dibaca"))
		return
	}
	contentType := http.DetectContentType(head[:n])
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		dto.RespondError(c, domain.BadRequest("File tidak dapat dibaca"))
		return
	}

	photo, err := h.uc.UploadPhoto(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		c.Param("id"),
		f,
		fileHeader.Size,
		contentType,
	)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusCreated, dto.NewPhotoResponse(photo), "Foto berhasil diunggah")
}

func (h *PropertyHandler) DeletePhoto(c *gin.Context) {
	if err := h.uc.RemovePhoto(
		c.Request.Context(),
		c.GetString(middleware.ContextUserID),
		c.Param("id"),
		c.Param("photoId"),
	); err != nil {
		dto.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
