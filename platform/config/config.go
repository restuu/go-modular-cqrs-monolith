package config

import (
	"fmt"
	"os"
)

// Config holds all configuration loaded from environment variables.
// No hardcoded values — everything comes from the environment.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Auth     AuthConfig
	AI       AIConfig
	Internal InternalConfig
}

type AppConfig struct {
	Port string
	Env  string // "development" | "production"
}

type DatabaseConfig struct {
	DSN string // full postgres connection string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret                string
	AccessTokenTTLMinutes int
	RefreshTokenTTLDays   int
}

// AuthConfig selects how caller identity is established for protected routes.
//
//   - Mode "local": this service issues and verifies JWTs itself; the auth
//     module is constructed and mounts /api/v1/auth/*.
//   - Mode "gateway": an upstream gateway (e.g. Kong) validates the JWT and
//     forwards the parsed claims as headers; this service trusts them and
//     never sees a raw JWT. The auth module is not constructed.
//
// Swapping modes is the detachability proof for the auth module: every other
// module reads identity only through platform/httpx/middleware's Locals contract,
// so neither mode requires any domain-module change.
type AuthConfig struct {
	Mode         string
	UserIDHeader string // gateway mode only: header carrying the authenticated user's UUID
	RoleHeader   string // gateway mode only: header carrying the user's role
}

type AIConfig struct {
	BaseURL string // e.g. http://ai-service:8000
}

type InternalConfig struct {
	// APISecret guards /internal/* endpoints with a shared header (X-Internal-Secret).
	// Empty in dev; required non-empty in production.
	APISecret string
}

// Load reads all required config from environment.
// Returns an error for any missing required variable — the caller decides how to handle it
// (main.go calls fatal(), tests can assert the error directly).
func Load() (*Config, error) {
	authMode := getEnvOrDefault("AUTH_MODE", "local")
	jwtSecret := getEnvOrDefault("JWT_SECRET", "")
	if authMode == "local" && jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required when AUTH_MODE=local")
	}

	dsn, err := requireEnv("DATABASE_DSN")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Port: getEnvOrDefault("APP_PORT", "8080"),
			Env:  getEnvOrDefault("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			DSN: dsn,
		},
		Redis: RedisConfig{
			Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
			Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		},
		JWT: JWTConfig{
			Secret:                jwtSecret,
			AccessTokenTTLMinutes: 15,
			RefreshTokenTTLDays:   7,
		},
		Auth: AuthConfig{
			Mode:         authMode,
			UserIDHeader: getEnvOrDefault("AUTH_GATEWAY_USER_ID_HEADER", "X-User-Id"),
			RoleHeader:   getEnvOrDefault("AUTH_GATEWAY_ROLE_HEADER", "X-User-Role"),
		},
		AI: AIConfig{
			BaseURL: getEnvOrDefault("AI_SERVICE_URL", "http://localhost:8000"),
		},
		Internal: InternalConfig{
			APISecret: getEnvOrDefault("INTERNAL_API_SECRET", ""),
		},
	}
	return cfg, nil
}

// requireEnv returns the value of key, or an error if it is unset or empty.
// Returning an error (rather than panicking) keeps Load() the single source of truth
// for startup failures — the wiring layer (main.go) decides how to handle them.
func requireEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}
	return val, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
