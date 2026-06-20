package router

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// healthEnvelope mirrors the production response shape for assertions.
type healthEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

type healthData struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func doHealthGet(t *testing.T, app *fiber.App) (int, healthData) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var env healthEnvelope
	require.NoError(t, json.Unmarshal(body, &env))
	var data healthData
	require.NoError(t, json.Unmarshal(env.Data, &data))
	return resp.StatusCode, data
}

func TestHealthz_no_probes_returns_200(t *testing.T) {
	app := fiber.New()
	registerHealth(app, nil)

	code, data := doHealthGet(t, app)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "ok", data.Status)
	assert.Empty(t, data.Checks)
}

func TestHealthz_all_probes_pass_returns_200(t *testing.T) {
	app := fiber.New()
	registerHealth(app, map[string]func(ctx context.Context) error{
		"postgres": func(ctx context.Context) error { return nil },
		"redis":    func(ctx context.Context) error { return nil },
	})

	code, data := doHealthGet(t, app)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "ok", data.Status)
	assert.Equal(t, "ok", data.Checks["postgres"])
	assert.Equal(t, "ok", data.Checks["redis"])
}

func TestHealthz_failing_probe_returns_503_degraded(t *testing.T) {
	app := fiber.New()
	registerHealth(app, map[string]func(ctx context.Context) error{
		"postgres": func(ctx context.Context) error { return nil },
		"redis":    func(ctx context.Context) error { return errors.New("connection refused") },
	})

	code, data := doHealthGet(t, app)

	assert.Equal(t, http.StatusServiceUnavailable, code)
	assert.Equal(t, "degraded", data.Status)
	assert.Equal(t, "ok", data.Checks["postgres"])
	assert.Contains(t, data.Checks["redis"], "connection refused")
}
