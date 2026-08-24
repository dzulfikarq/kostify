package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Connect(ctx context.Context, dbURL string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	for attempt := 1; attempt <= 5; attempt++ {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Warn),
			NowFunc: func() time.Time { return time.Now().UTC() },
		})
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	if err != nil {
		return nil, fmt.Errorf("koneksi postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return sqlDB.PingContext(pingCtx)
}

func Migrate(migrateURL, dir string, up bool, steps int) error {
	m, err := migrate.New("file://"+dir, migrateURL)
	if err != nil {
		return fmt.Errorf("init migrasi: %w", err)
	}
	defer m.Close()

	if up {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate up: %w", err)
		}
		return nil
	}
	if steps < 0 {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate down: %w", err)
		}
		return nil
	}
	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down %d: %w", steps, err)
	}
	return nil
}
