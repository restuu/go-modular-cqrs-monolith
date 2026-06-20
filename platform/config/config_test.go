package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-modular-cqrs-monolith/platform/config"
)

// setRequiredBaseEnv sets the env vars Load() always requires regardless of
// auth mode, so each test only needs to set the auth-specific vars under test.
func setRequiredBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/db?sslmode=disable")
}

func TestLoad_AuthMode(t *testing.T) {
	tests := []struct {
		name          string
		authMode      string // empty means unset
		jwtSecret     string // empty means unset
		wantErr       bool
		wantMode      string
		wantJWTSecret string
	}{
		{
			name:          "defaults_to_local_when_AUTH_MODE_unset_and_secret_present",
			authMode:      "",
			jwtSecret:     "a-very-secret-value",
			wantErr:       false,
			wantMode:      "local",
			wantJWTSecret: "a-very-secret-value",
		},
		{
			name:      "local_mode_without_JWT_SECRET_returns_error",
			authMode:  "local",
			jwtSecret: "",
			wantErr:   true,
		},
		{
			name:          "gateway_mode_without_JWT_SECRET_succeeds",
			authMode:      "gateway",
			jwtSecret:     "",
			wantErr:       false,
			wantMode:      "gateway",
			wantJWTSecret: "",
		},
		{
			name:          "gateway_mode_with_JWT_SECRET_still_succeeds",
			authMode:      "gateway",
			jwtSecret:     "unused-but-present",
			wantErr:       false,
			wantMode:      "gateway",
			wantJWTSecret: "unused-but-present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredBaseEnv(t)
			if tt.authMode != "" {
				t.Setenv("AUTH_MODE", tt.authMode)
			}
			if tt.jwtSecret != "" {
				t.Setenv("JWT_SECRET", tt.jwtSecret)
			}

			cfg, err := config.Load()

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMode, cfg.Auth.Mode)
			assert.Equal(t, tt.wantJWTSecret, cfg.JWT.Secret)
		})
	}
}

func TestLoad_AuthGatewayHeaders(t *testing.T) {
	t.Run("defaults_to_X-User-Id_and_X-User-Role_when_unset", func(t *testing.T) {
		setRequiredBaseEnv(t)
		t.Setenv("AUTH_MODE", "gateway")

		cfg, err := config.Load()

		require.NoError(t, err)
		assert.Equal(t, "X-User-Id", cfg.Auth.UserIDHeader)
		assert.Equal(t, "X-User-Role", cfg.Auth.RoleHeader)
	})

	t.Run("uses_custom_header_names_when_set", func(t *testing.T) {
		setRequiredBaseEnv(t)
		t.Setenv("AUTH_MODE", "gateway")
		t.Setenv("AUTH_GATEWAY_USER_ID_HEADER", "X-Consumer-Id")
		t.Setenv("AUTH_GATEWAY_ROLE_HEADER", "X-Consumer-Role")

		cfg, err := config.Load()

		require.NoError(t, err)
		assert.Equal(t, "X-Consumer-Id", cfg.Auth.UserIDHeader)
		assert.Equal(t, "X-Consumer-Role", cfg.Auth.RoleHeader)
	})
}
