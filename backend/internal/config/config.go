package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv          string
	AppPort         string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	RedisAddr       string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CookieSecure    bool
	CookieDomain    string
	FrontendOrigins []string
	AutoMigrate     bool
	MigrationsDir   string
	BcryptCost      int
	MinIOEndpoint   string
	MinIOAccessKey  string
	MinIOSecretKey  string
	MinioBucket     string
	MinioSecure     bool
	MinioPublicURL  string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:          env("APP_ENV", "development"),
		AppPort:         env("APP_PORT", "8080"),
		DBHost:          env("DB_HOST", "localhost"),
		DBPort:          env("DB_PORT", "5432"),
		DBUser:          env("DB_USER", "kostify"),
		DBPassword:      env("DB_PASSWORD", "kostify"),
		DBName:          env("DB_NAME", "kostify"),
		RedisAddr:       env("REDIS_ADDR", "localhost:6379"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AccessTokenTTL:  envDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: envDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		CookieSecure:    envBool("COOKIE_SECURE", false),
		CookieDomain:    os.Getenv("COOKIE_DOMAIN"),
		FrontendOrigins: splitCSV(env("FRONTEND_ORIGIN", "http://localhost:5173")),
		AutoMigrate:     envBool("AUTO_MIGRATE", false),
		MigrationsDir:   env("MIGRATIONS_DIR", "migrations"),
		BcryptCost:      envInt("BCRYPT_COST", 12),
		MinIOEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinIOAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinIOSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:     os.Getenv("MINIO_BUCKET"),
		MinioSecure:     envBool("MINIO_SECURE", false),
		MinioPublicURL:  os.Getenv("MINIO_PUBLIC_URL"),
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET minimal 32 karakter")
	}
	if cfg.MinIOEndpoint == "" || cfg.MinioBucket == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT dan MINIO_BUCKET wajib diisi")
	}
	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func (c *Config) MigrateURL() string {
	return fmt.Sprintf(
		"pgx5://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func (c *Config) IsProduction() bool { return c.AppEnv == "production" }

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := strings.ToLower(os.Getenv(key))
	switch v {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return def
	}
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
