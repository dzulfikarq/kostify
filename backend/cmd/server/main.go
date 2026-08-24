package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	infrapostgres "github.com/dzulfikarq/kostify/backend/internal/infrastructure/postgres"
	infraredis "github.com/dzulfikarq/kostify/backend/internal/infrastructure/redis"
	httpapi "github.com/dzulfikarq/kostify/backend/internal/interfaces/http"

	"github.com/dzulfikarq/kostify/backend/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	autoMigrate := flag.Bool("automigrate", false, "jalankan migration sebelum serve")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config tidak valid", "err", err.Error())
		os.Exit(1)
	}

	ctx := context.Background()

	db, err := infrapostgres.Connect(ctx, cfg.DatabaseURL())
	if err != nil {
		slog.Error("gagal konek postgres", "err", err.Error())
		os.Exit(1)
	}

	if *autoMigrate || cfg.AutoMigrate {
		if err := infrapostgres.Migrate(cfg.MigrateURL(), cfg.MigrationsDir, true, -1); err != nil {
			slog.Error("migration gagal", "err", err.Error())
			os.Exit(1)
		}
		slog.Info("migration selesai")
	}

	rdb := infraredis.NewClient(cfg.RedisAddr)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	err = infraredis.Ping(pingCtx, rdb)
	cancel()
	if err != nil {
		slog.Error("gagal konek redis", "err", err.Error())
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           httpapi.NewRouter(cfg, db, rdb),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("server berjalan", "port", cfg.AppPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server berhenti", "err", err.Error())
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutdown dimulai")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown gagal", "err", err.Error())
	}

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	_ = rdb.Close()
	slog.Info("shutdown selesai")
}
