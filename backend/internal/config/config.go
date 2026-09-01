package config

import (
	"os"
	"strconv"
	"time"

	"errors"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string

	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	RefreshTokenLen int

	SeedSuperAdmin     bool
	SuperAdminName     string
	SuperAdminEmail    string
	SuperAdminPassword string
	SuperAdminRole     string

	UploadDir        string
	MaxUploadSizeMB  int64
	RateLimitPerIP   int
	RateLimitWindowS int

	AllowedOrigins []string
	TestDBURL      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("backend/.env")

	cfg := &Config{
		Environment:        os.Getenv("APP_ENV"),
		Port:               orDefault("PORT", "8080"),
		DBHost:             orDefault("DB_HOST", "localhost"),
		DBPort:             orDefault("DB_PORT", "5432"),
		DBUser:             orDefault("DB_USER", "emplyra"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             orDefault("DB_NAME", "emplyra"),
		DBSSLMode:          orDefault("DB_SSLMODE", "disable"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AccessTokenTTL:     mustDuration("ACCESS_TOKEN_TTL", 2*time.Hour),
		RefreshTokenTTL:    mustDuration("REFRESH_TOKEN_TTL", 24*7*time.Hour),
		RefreshTokenLen:    mustInt("REFRESH_TOKEN_LEN", 64),
		SeedSuperAdmin:     mustBool("SEED_SUPER_ADMIN", true),
		SuperAdminName:     orDefault("SUPER_ADMIN_NAME", "Super Admin"),
		SuperAdminEmail:    orDefault("SUPER_ADMIN_EMAIL", "admin@emplyra.local"),
		SuperAdminPassword: orDefault("SUPER_ADMIN_PASSWORD", "ChangeMe123!"),
		SuperAdminRole:     orDefault("SUPER_ADMIN_ROLE", "SUPER_ADMIN"),
		UploadDir:          orDefault("UPLOAD_DIR", "./uploads"),
		MaxUploadSizeMB:    mustInt64("MAX_UPLOAD_MB", 10),
		RateLimitPerIP:     mustInt("RATE_LIMIT_PER_IP", 60),
		RateLimitWindowS:   mustInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		AllowedOrigins:     splitComma(orDefault("CORS_ORIGINS", "*")),
		TestDBURL:          os.Getenv("TEST_DB_URL"),
	}

	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required (set it in .env)")
	}
	if cfg.UploadDir == "" {
		cfg.UploadDir = "./uploads"
	}
	return cfg, nil
}

func orDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func mustInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func mustBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func mustDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if val := trimSpace(s[start:i]); val != "" {
				out = append(out, val)
			}
			start = i + 1
		}
	}
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
