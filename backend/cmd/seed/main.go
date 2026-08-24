package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/dzulfikarq/kostify/backend/internal/config"
	infrapostgres "github.com/dzulfikarq/kostify/backend/internal/infrastructure/postgres"
	"github.com/dzulfikarq/kostify/backend/internal/domain"
	"github.com/dzulfikarq/kostify/backend/internal/pkg/password"
)

var seeds = []struct {
	id       string
	name     string
	email    string
	password string
	role     domain.Role
}{
	{"00000000-0000-0000-0000-000000000001", "Super Admin", "admin@kostify.test", "Admin123!", domain.RoleSuperAdmin},
	{"00000000-0000-0000-0000-000000000002", "Pemilik Kost", "owner@kostify.test", "Owner123!", domain.RoleOwner},
	{"00000000-0000-0000-0000-000000000003", "Calon Penghuni", "tenant@kostify.test", "Tenant123!", domain.RoleTenant},
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config tidak valid: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := infrapostgres.Connect(ctx, cfg.DatabaseURL())
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal konek postgres: %v\n", err)
		os.Exit(1)
	}
	repo := infrapostgres.NewUserRepo(db)

	for _, s := range seeds {
		if _, err := repo.FindByEmail(ctx, s.email); err == nil {
			slog.Info("skip (sudah ada)", "email", s.email)
			continue
		}
		hash, err := password.Hash(s.password, cfg.BcryptCost)
		if err != nil {
			fatal("hash password %s: %v", s.email, err)
		}
		user := &domain.User{
			ID:           s.id,
			Name:         s.name,
			Email:        s.email,
			PasswordHash: hash,
			Role:         s.role,
		}
		if err := db.WithContext(ctx).Exec(
			`INSERT INTO users (id, name, email, password_hash, role) VALUES (?, ?, ?, ?, ?::user_role)`,
			user.ID, user.Name, user.Email, user.PasswordHash, string(user.Role),
		).Error; err != nil {
			fatal("insert %s: %v", s.email, err)
		}
		slog.Info("seed dibuat", "email", s.email, "role", string(s.role))
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
