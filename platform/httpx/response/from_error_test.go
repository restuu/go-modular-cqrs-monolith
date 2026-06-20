package response_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-modular-cqrs-monolith/platform/errcode"
	"go-modular-cqrs-monolith/platform/httpx/response"
)

// makeApp wires a minimal Fiber app that calls FromError with the given error.
func makeApp(t *testing.T, err error, logBuf *bytes.Buffer) *fiber.App {
	t.Helper()
	logger := slog.New(slog.NewJSONHandler(logBuf, nil))
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return response.FromError(c, logger, err)
	})
	return app
}

func doGet(t *testing.T, app *fiber.App) (int, Envelope) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var env Envelope
	require.NoError(t, json.Unmarshal(body, &env))
	return resp.StatusCode, env
}

// Envelope mirrors response.Envelope for decoding.
type Envelope struct {
	Success bool    `json:"success"`
	Error   *APIErr `json:"error"`
}

type APIErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestFromError_AppError_NotFound(t *testing.T) {
	err := errcode.New(errcode.Code("ARTICLE_NOT_FOUND"), "article not found")
	status, env := doGet(t, makeApp(t, err, &bytes.Buffer{}))

	assert.Equal(t, fiber.StatusNotFound, status)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "ARTICLE_NOT_FOUND", env.Error.Code)
	assert.Equal(t, "article not found", env.Error.Message)
}

func TestFromError_AppError_Duplicate(t *testing.T) {
	err := errcode.New(errcode.Code("TAG_DUPLICATE"), "tag already exists")
	status, env := doGet(t, makeApp(t, err, &bytes.Buffer{}))

	assert.Equal(t, fiber.StatusConflict, status)
	assert.Equal(t, "TAG_DUPLICATE", env.Error.Code)
}

func TestFromError_AppError_Taken(t *testing.T) {
	err := errcode.New(errcode.Code("USERNAME_TAKEN"), "username taken")
	status, _ := doGet(t, makeApp(t, err, &bytes.Buffer{}))

	assert.Equal(t, fiber.StatusConflict, status)
}

func TestFromError_AppError_AlreadyBookmarked(t *testing.T) {
	err := errcode.New(errcode.Code("ALREADY_BOOKMARKED"), "already bookmarked")
	status, env := doGet(t, makeApp(t, err, &bytes.Buffer{}))

	assert.Equal(t, fiber.StatusConflict, status)
	assert.Equal(t, "ALREADY_BOOKMARKED", env.Error.Code)
}

func TestFromError_AppError_Unauthorized(t *testing.T) {
	err := errcode.New(errcode.Unauthorized, "invalid credentials")
	status, env := doGet(t, makeApp(t, err, &bytes.Buffer{}))

	assert.Equal(t, fiber.StatusUnauthorized, status)
	assert.Equal(t, "UNAUTHORIZED", env.Error.Code)
}

func TestFromError_AppError_Forbidden(t *testing.T) {
	err := errcode.New(errcode.Forbidden, "forbidden")
	status, _ := doGet(t, makeApp(t, err, &bytes.Buffer{}))

	assert.Equal(t, fiber.StatusForbidden, status)
}

func TestFromError_AppError_ValidationCodes(t *testing.T) {
	codes := []errcode.Code{
		errcode.ValidationError,
		errcode.InvalidBody,
		errcode.InvalidQuery,
		errcode.InvalidID,
		errcode.InvalidPagination,
	}
	for _, code := range codes {
		code := code
		t.Run(string(code), func(t *testing.T) {
			err := errcode.New(code, "bad request")
			status, _ := doGet(t, makeApp(t, err, &bytes.Buffer{}))
			assert.Equal(t, fiber.StatusBadRequest, status)
		})
	}
}

func TestFromError_UnknownError_Returns500AndLogs(t *testing.T) {
	var logBuf bytes.Buffer
	unknown := errors.New("unexpected db connection lost")
	status, env := doGet(t, makeApp(t, unknown, &logBuf))

	assert.Equal(t, fiber.StatusInternalServerError, status)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "INTERNAL_ERROR", env.Error.Code)
	assert.Equal(t, "an unexpected error occurred", env.Error.Message)

	// slog must have emitted a JSON log line containing the error text
	assert.Contains(t, logBuf.String(), "unexpected db connection lost")
	assert.Contains(t, logBuf.String(), "unhandled error")
}

func TestFromError_WrappedAppError_ResolvesCorrectly(t *testing.T) {
	inner := errcode.New(errcode.Code("SOURCE_NOT_FOUND"), "source not found")
	wrapped := errcode.Wrap(inner, errcode.Code("SOURCE_NOT_FOUND"), "source not found")
	status, env := doGet(t, makeApp(t, wrapped, &bytes.Buffer{}))

	assert.Equal(t, fiber.StatusNotFound, status)
	assert.Equal(t, "SOURCE_NOT_FOUND", env.Error.Code)
}
