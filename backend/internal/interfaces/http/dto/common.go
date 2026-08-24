package dto

import (
	"errors"
	"log/slog"

	"github.com/dzulfikarq/kostify/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func OK(c *gin.Context, status int, data any, message string) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
		"message": message,
	})
}

func RespondError(c *gin.Context, err error) {
	requestID, _ := c.Get("request_id")

	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		details := make([]gin.H, 0, len(ve.Details))
		for _, d := range ve.Details {
			details = append(details, gin.H{"field": d.Field, "message": d.Message})
		}
		c.AbortWithStatusJSON(422, gin.H{
			"success": false,
			"error": gin.H{
				"code":    domain.CodeValidation,
				"message": "Input tidak valid",
				"details": details,
			},
			"request_id": requestID,
		})
		return
	}

	var ae *domain.APIError
	if errors.As(err, &ae) {
		c.AbortWithStatusJSON(ae.Status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    ae.Code,
				"message": ae.Message,
			},
			"request_id": requestID,
		})
		return
	}

	slog.Error("internal error", "err", err.Error(), "request_id", requestID)
	c.AbortWithStatusJSON(500, gin.H{
		"success": false,
		"error": gin.H{
			"code":    domain.CodeInternal,
			"message": "Terjadi kesalahan pada server",
		},
		"request_id": requestID,
	})
}
