package metrics_test

import (
	"bufio"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-modular-cqrs-monolith/platform/metrics"
)

func TestHandler_serves_prometheus_text_format(t *testing.T) {
	reg, rec := metrics.New()
	rec.ObserveRequest("GET", "/api/v1/articles", 200, 5*time.Millisecond)

	app := fiber.New()
	app.Get("/internal/metrics", metrics.Handler(reg))

	req := httptest.NewRequest("GET", "/internal/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	text := string(body)

	assert.Contains(t, text, "http_requests_total")
	assert.Contains(t, text, "http_request_duration_seconds")
	assert.Contains(t, text, "go_goroutines")
}

func TestHandler_returns_valid_text_exposition(t *testing.T) {
	reg, _ := metrics.New()

	app := fiber.New()
	app.Get("/metrics", metrics.Handler(reg))

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Every non-comment line must be parseable as "name{...} value" or "name value"
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.Fields(line)
		assert.GreaterOrEqual(t, len(parts), 2, "metric line should have at least name and value: %q", line)
	}
	require.NoError(t, scanner.Err())
}
