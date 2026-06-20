package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"go-modular-cqrs-monolith/platform/httpx/middleware"
	"go-modular-cqrs-monolith/platform/logging"
)

// appWithRequestID builds a minimal Fiber app wired with RequestID middleware.
// The single test route echoes the locals request_id and context request_id
// as response headers so tests can assert them without reaching into internals.
func appWithRequestID() *fiber.App {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		// Echo the Locals value.
		if id, ok := c.Locals(middleware.LocalsRequestID).(uuid.UUID); ok {
			c.Set("X-Locals-ID", id.String())
		}
		// Echo the context value set by WithRequestID.
		if id, ok := logging.RequestIDFromContext(c.Context()); ok {
			c.Set("X-Context-ID", id.String())
		}
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

type RequestIDMiddlewareSuite struct {
	suite.Suite
	app *fiber.App
}

func (s *RequestIDMiddlewareSuite) SetupTest() {
	s.app = appWithRequestID()
}

func TestRequestIDMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(RequestIDMiddlewareSuite))
}

// do issues a GET / with optional headers and returns (statusCode, responseHeaders).
// The response body is consumed and closed internally.
func (s *RequestIDMiddlewareSuite) do(headers map[string]string) (int, http.Header) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.app.Test(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	return resp.StatusCode, resp.Header
}

func (s *RequestIDMiddlewareSuite) TestGeneratesUUID_when_header_absent() {
	_, headers := s.do(nil)

	headerID := headers.Get("X-Request-ID")
	s.NotEmpty(headerID)
	_, err := uuid.Parse(headerID)
	s.NoError(err, "generated ID must be a valid UUID")
}

func (s *RequestIDMiddlewareSuite) TestPreservesInboundUUID_when_header_is_valid() {
	id := uuid.New()
	_, headers := s.do(map[string]string{"X-Request-ID": id.String()})

	s.Equal(id.String(), headers.Get("X-Request-ID"))
	s.Equal(id.String(), headers.Get("X-Locals-ID"))
	s.Equal(id.String(), headers.Get("X-Context-ID"))
}

func (s *RequestIDMiddlewareSuite) TestRegeneratesUUID_when_header_is_malformed() {
	_, headers := s.do(map[string]string{"X-Request-ID": "not-a-uuid"})

	headerID := headers.Get("X-Request-ID")
	s.NotEmpty(headerID)
	s.NotEqual("not-a-uuid", headerID)
	_, err := uuid.Parse(headerID)
	s.NoError(err, "regenerated ID must be a valid UUID")
}

func (s *RequestIDMiddlewareSuite) TestEchosIDInResponseHeader() {
	_, headers := s.do(nil)
	s.NotEmpty(headers.Get("X-Request-ID"))
}

func (s *RequestIDMiddlewareSuite) TestLocalsAndContextCarrySameID() {
	_, headers := s.do(nil)

	localsID := headers.Get("X-Locals-ID")
	contextID := headers.Get("X-Context-ID")
	responseID := headers.Get("X-Request-ID")

	s.NotEmpty(localsID)
	s.Equal(localsID, contextID, "Locals and context must hold the same ID")
	s.Equal(localsID, responseID, "response header must match locals")
}
