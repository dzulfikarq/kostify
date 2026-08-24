package http

import (
	"log/slog"
	"net/http"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/dzulfikarq/kostify/backend/internal/application"
	"github.com/dzulfikarq/kostify/backend/internal/config"
	"github.com/dzulfikarq/kostify/backend/internal/domain"
	infrapostgres "github.com/dzulfikarq/kostify/backend/internal/infrastructure/postgres"
	infraredis "github.com/dzulfikarq/kostify/backend/internal/infrastructure/redis"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/dto"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/handler"
	"github.com/dzulfikarq/kostify/backend/internal/interfaces/http/middleware"
	"github.com/dzulfikarq/kostify/backend/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, db *gorm.DB, rdb *goredis.Client) http.Handler {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	userRepo := infrapostgres.NewUserRepo(db)
	tokenStore := infraredis.NewTokenStore(rdb, cfg.RefreshTokenTTL)
	signer := jwt.NewSigner([]byte(cfg.JWTSecret))
	authUC := application.NewAuthUsecase(userRepo, tokenStore, signer, cfg.AccessTokenTTL, cfg.BcryptCost)
	authH := handler.NewAuthHandler(authUC, cfg)

	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.SecurityHeaders(),
		middleware.CORS(cfg.FrontendOrigins),
	)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := infrapostgres.Ping(ctx, db); err != nil {
			slog.Error("readyz: postgres tidak siap", "err", err.Error())
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "postgres unavailable"})
			return
		}
		if err := infraredis.Ping(ctx, rdb); err != nil {
			slog.Error("readyz: redis tidak siap", "err", err.Error())
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "redis unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", middleware.LoginRateLimit(rdb), authH.Login)
		auth.POST("/refresh", authH.Refresh)
	}
	authed := auth.Group("", middleware.Auth(signer))
	{
		authed.GET("/me", authH.Me)
		authed.POST("/logout", middleware.CSRF(), authH.Logout)
	}

	v1.GET("/admin/ping",
		middleware.Auth(signer),
		middleware.RequireRole(domain.RoleSuperAdmin),
		func(c *gin.Context) {
			dto.OK(c, http.StatusOK, gin.H{"pong": true}, "Success")
		},
	)

	return r
}
