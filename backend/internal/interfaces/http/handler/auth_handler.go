package handler

import (
	"net/http"

	"github.com/dzulfikarq/kostify/backend/internal/application"
	"github.com/dzulfikarq/kostify/backend/internal/config"
	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/middleware"
	"github.com/dzulfikarq/kostify/backend/internal/pkg/random"

	"github.com/gin-gonic/gin"
)

const (
	cookieAccess  = "access_token"
	cookieRefresh = "refresh_token"
)

type AuthHandler struct {
	uc           *application.AuthUsecase
	cfg          *config.Config
}

func NewAuthHandler(uc *application.AuthUsecase, cfg *config.Config) *AuthHandler {
	return &AuthHandler{uc: uc, cfg: cfg}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}

	user, err := h.uc.Register(c.Request.Context(), application.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusCreated, dto.NewUserResponse(user), "Registrasi berhasil")
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, domain.BadRequest("Body request tidak valid"))
		return
	}

	result, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	csrf := random.String(16)
	h.setCookies(c, result.AccessToken, result.RefreshToken, csrf)
	dto.OK(c, http.StatusOK, &dto.AuthData{
		User:      dto.NewUserResponse(result.User),
		CSRFToken: csrf,
	}, "Login berhasil")
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refresh, _ := c.Cookie(cookieRefresh)
	result, err := h.uc.Refresh(c.Request.Context(), refresh)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	csrf := random.String(16)
	h.setCookies(c, result.AccessToken, result.RefreshToken, csrf)
	dto.OK(c, http.StatusOK, &dto.AuthData{
		User:      dto.NewUserResponse(result.User),
		CSRFToken: csrf,
	}, "Token diperbarui")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refresh, _ := c.Cookie(cookieRefresh)
	if err := h.uc.Logout(c.Request.Context(), refresh); err != nil {
		dto.RespondError(c, err)
		return
	}
	h.clearCookies(c)
	dto.OK(c, http.StatusOK, nil, "Logout berhasil")
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserID)
	user, err := h.uc.Me(c.Request.Context(), userID)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.OK(c, http.StatusOK, dto.NewUserResponse(user), "Success")
}

func (h *AuthHandler) setCookies(c *gin.Context, access, refresh, csrf string) {
	secure := h.cfg.CookieSecure
	domainAttr := h.cfg.CookieDomain

	set := func(name, value string, maxAge int, httpOnly bool) {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(name, value, maxAge, "/", domainAttr, secure, httpOnly)
	}
	set(cookieAccess, access, int(h.cfg.AccessTokenTTL.Seconds()), true)
	set(cookieRefresh, refresh, int(h.cfg.RefreshTokenTTL.Seconds()), true)
	set("csrf_token", csrf, int(h.cfg.RefreshTokenTTL.Seconds()), false)
}

func (h *AuthHandler) clearCookies(c *gin.Context) {
	set := func(name string, httpOnly bool) {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(name, "", -1, "/", h.cfg.CookieDomain, h.cfg.CookieSecure, httpOnly)
	}
	set(cookieAccess, true)
	set(cookieRefresh, true)
	set("csrf_token", false)
}
