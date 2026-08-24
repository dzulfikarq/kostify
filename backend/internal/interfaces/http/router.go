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
	mailerpkg "github.com/dzulfikarq/kostify/backend/internal/pkg/mailer"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, db *gorm.DB, rdb *goredis.Client, storage domain.ObjectStorage) http.Handler {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	userRepo := infrapostgres.NewUserRepo(db)
	propertyRepo := infrapostgres.NewPropertyRepo(db)
	roomRepo := infrapostgres.NewRoomRepo(db)
	bookingRepo := infrapostgres.NewBookingRepo(db)
	tokenStore := infraredis.NewTokenStore(rdb, cfg.RefreshTokenTTL)
	signer := jwt.NewSigner([]byte(cfg.JWTSecret))

	authUC := application.NewAuthUsecase(userRepo, tokenStore, signer, cfg.AccessTokenTTL, cfg.BcryptCost)
	propertyUC := application.NewPropertyUsecase(propertyRepo, roomRepo, storage)
	roomUC := application.NewRoomUsecase(roomRepo, propertyRepo)
	notificationRepo := infrapostgres.NewNotificationRepo(db)
	mailer := mailerpkg.New(cfg.SMTPHost, cfg.SMTPPort, cfg.MailFrom)
	bookingUC := application.NewBookingUsecase(bookingRepo, notificationRepo, userRepo, mailer)

	reviewRepo := infrapostgres.NewReviewRepo(db)
	wishlistRepo := infrapostgres.NewWishlistRepo(db)
	dashboardRepo := infrapostgres.NewDashboardRepo(db)
	reviewUC := application.NewReviewUsecase(reviewRepo, bookingRepo)
	wishlistUC := application.NewWishlistUsecase(wishlistRepo, propertyRepo)
	dashboardUC := application.NewDashboardUsecase(dashboardRepo)

	authH := handler.NewAuthHandler(authUC, cfg)
	propertyH := handler.NewPropertyHandler(propertyUC)
	roomH := handler.NewRoomHandler(roomUC)
	bookingH := handler.NewBookingHandler(bookingUC)
	notifikasiH := handler.NewNotificationHandler(notificationRepo)
	reviewH := handler.NewReviewHandler(reviewUC, wishlistUC, dashboardUC)

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

	propsPublic := v1.Group("/properties")
	{
		propsPublic.GET("", propertyH.ListPublic)
		propsPublic.GET("/:id", middleware.OptionalAuth(signer), propertyH.GetDetail)
		propsPublic.GET("/:id/rooms", middleware.OptionalAuth(signer), roomH.List)
		propsPublic.GET("/:id/reviews", reviewH.ListByProperty)
	}

	owner := v1.Group("", middleware.Auth(signer), middleware.RequireRole(domain.RoleOwner))
	{
		owner.GET("/properties/owner", propertyH.ListMine)
		owner.POST("/properties", propertyH.Create)
		owner.PUT("/properties/:id", propertyH.Update)
		owner.POST("/properties/:id/photos", propertyH.UploadPhoto)
		owner.DELETE("/properties/:id/photos/:photoId", propertyH.DeletePhoto)
		owner.POST("/properties/:id/submit", propertyH.Submit)

		owner.POST("/properties/:id/rooms", roomH.Create)
		owner.PUT("/rooms/:id", roomH.Update)
		owner.DELETE("/rooms/:id", roomH.Delete)
		owner.PATCH("/rooms/:id/status", roomH.UpdateStatus)
	}

	admin := v1.Group("", middleware.Auth(signer), middleware.RequireRole(domain.RoleSuperAdmin))
	{
		admin.POST("/properties/:id/approve", propertyH.Approve)
		admin.POST("/properties/:id/reject", propertyH.Reject)
		admin.DELETE("/properties/:id", propertyH.AdminDelete)
	}

	tenant := v1.Group("", middleware.Auth(signer), middleware.RequireRole(domain.RoleTenant))
	{
		tenant.POST("/bookings", bookingH.Create)
		tenant.GET("/bookings/me", bookingH.ListMine)
		tenant.PUT("/bookings/:id/checkin", bookingH.CheckIn)
		tenant.PUT("/bookings/:id/checkout", bookingH.CheckOut)
		tenant.PUT("/bookings/:id/cancel", bookingH.Cancel)
		tenant.POST("/bookings/:id/reviews", reviewH.Create)

		tenant.GET("/wishlist", reviewH.WishlistList)
		tenant.POST("/wishlist", reviewH.WishlistAdd)
		tenant.DELETE("/wishlist/:propertyId", reviewH.WishlistRemove)
	}

	v1.GET("/dashboard", middleware.Auth(signer), reviewH.Dashboard)

	bookingOwner := v1.Group("", middleware.Auth(signer), middleware.RequireRole(domain.RoleOwner))
	{
		bookingOwner.GET("/bookings/owner", bookingH.ListOwner)
		bookingOwner.PUT("/bookings/:id/approve", bookingH.Approve)
		bookingOwner.PUT("/bookings/:id/reject", bookingH.Reject)
		bookingOwner.PUT("/bookings/:id/confirm", bookingH.Confirm)
	}

	authedNotif := v1.Group("", middleware.Auth(signer))
	{
		authedNotif.GET("/notifications", notifikasiH.List)
		authedNotif.PUT("/notifications/read-all", notifikasiH.MarkAllRead)
		authedNotif.PUT("/notifications/:id/read", notifikasiH.MarkRead)
	}

	return r
}
