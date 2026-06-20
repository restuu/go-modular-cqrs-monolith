package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"go-modular-cqrs-monolith/platform/httpx/middleware"
)

type RequestLoggerMiddlewareSuite struct {
	suite.Suite
}

func TestRequestLoggerMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(RequestLoggerMiddlewareSuite))
}

// buildApp wires RequestID + RequestLogger in the correct order and registers
// a route that responds with the given status code.
func (s *RequestLoggerMiddlewareSuite) buildApp(statusCode int, buf *bytes.Buffer) *fiber.App {
	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	app := fiber.New(fiber.Config{})
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(logger))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendStatus(statusCode)
	})
	return app
}

// doGet issues a GET request and closes the response body, returning only the status.
func (s *RequestLoggerMiddlewareSuite) doGet(app *fiber.App, target string) int {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	resp, err := app.Test(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	return resp.StatusCode
}

// doGetWithHeaders issues a GET with extra request headers, closing the body.
func (s *RequestLoggerMiddlewareSuite) doGetWithHeaders(app *fiber.App, target string, headers map[string]string) int {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	return resp.StatusCode
}

func (s *RequestLoggerMiddlewareSuite) parseLog(buf *bytes.Buffer) map[string]any {
	s.T().Helper()
	var record map[string]any
	s.Require().NoError(json.Unmarshal(buf.Bytes(), &record))
	return record
}

func (s *RequestLoggerMiddlewareSuite) TestLogsStructuredFields_on_success() {
	buf := &bytes.Buffer{}
	app := s.buildApp(http.StatusOK, buf)

	status := s.doGet(app, "/test")
	s.Equal(http.StatusOK, status)

	record := s.parseLog(buf)
	s.Equal("GET", record["method"])
	s.Equal("/test", record["path"])
	s.Equal(float64(200), record["status"])
	s.NotEmpty(record["request_id"])
	_, err := uuid.Parse(record["request_id"].(string))
	s.NoError(err)
	_, hasDuration := record["duration_ms"]
	s.True(hasDuration)
}

func (s *RequestLoggerMiddlewareSuite) TestIncludesRequestID_from_prior_middleware() {
	buf := &bytes.Buffer{}
	app := s.buildApp(http.StatusOK, buf)

	inboundID := uuid.New()
	s.doGetWithHeaders(app, "/test", map[string]string{"X-Request-ID": inboundID.String()})

	record := s.parseLog(buf)
	s.Equal(inboundID.String(), record["request_id"])
}

func (s *RequestLoggerMiddlewareSuite) TestLogLevel_info_for_2xx() {
	buf := &bytes.Buffer{}
	app := s.buildApp(http.StatusOK, buf)
	s.doGet(app, "/test")

	record := s.parseLog(buf)
	s.Equal("INFO", record["level"])
}

func (s *RequestLoggerMiddlewareSuite) TestLogLevel_warn_for_4xx() {
	buf := &bytes.Buffer{}
	app := s.buildApp(http.StatusNotFound, buf)
	s.doGet(app, "/test")

	record := s.parseLog(buf)
	s.Equal("WARN", record["level"])
}

func (s *RequestLoggerMiddlewareSuite) TestLogLevel_error_for_5xx() {
	buf := &bytes.Buffer{}
	app := s.buildApp(http.StatusInternalServerError, buf)
	s.doGet(app, "/test")

	record := s.parseLog(buf)
	s.Equal("ERROR", record["level"])
}

func (s *RequestLoggerMiddlewareSuite) TestCapturesDuration() {
	buf := &bytes.Buffer{}
	app := s.buildApp(http.StatusOK, buf)
	s.doGet(app, "/test")

	record := s.parseLog(buf)
	durationMs, ok := record["duration_ms"].(float64)
	s.True(ok, "duration_ms must be present and numeric")
	s.GreaterOrEqual(durationMs, float64(0))
}
