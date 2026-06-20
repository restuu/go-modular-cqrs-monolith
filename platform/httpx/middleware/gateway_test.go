package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"go-modular-cqrs-monolith/platform/httpx/middleware"
)

// appWithGatewayClaims builds a minimal Fiber app wired with GatewayClaims,
// using the same header names a real Kong-style gateway would forward. The
// single test route echoes the locals user_id/user_role as response headers,
// and is also guarded by AdminOnly to prove the two middlewares share the
// same Locals contract regardless of which one set it.
func appWithGatewayClaims() *fiber.App {
	app := fiber.New()
	app.Get("/", middleware.GatewayClaims("X-User-Id", "X-User-Role"), func(c fiber.Ctx) error {
		if id, ok := middleware.GetUserID(c); ok {
			c.Set("X-Locals-ID", id.String())
		}
		if role, ok := c.Locals(middleware.LocalsRole).(string); ok {
			c.Set("X-Locals-Role", role)
		}
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/admin", middleware.GatewayClaims("X-User-Id", "X-User-Role"), middleware.AdminOnly(), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

type GatewayClaimsMiddlewareSuite struct {
	suite.Suite
	app *fiber.App
}

func (s *GatewayClaimsMiddlewareSuite) SetupTest() {
	s.app = appWithGatewayClaims()
}

func TestGatewayClaimsMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(GatewayClaimsMiddlewareSuite))
}

// do issues a GET to path with the given headers and returns (statusCode, responseHeaders).
func (s *GatewayClaimsMiddlewareSuite) do(path string, headers map[string]string) (int, http.Header) {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.app.Test(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	return resp.StatusCode, resp.Header
}

func (s *GatewayClaimsMiddlewareSuite) TestValidHeaders_SetsLocalsAndCallsNext() {
	userID := uuid.New()
	status, headers := s.do("/", map[string]string{
		"X-User-Id":   userID.String(),
		"X-User-Role": "admin",
	})

	s.Equal(fiber.StatusOK, status)
	s.Equal(userID.String(), headers.Get("X-Locals-ID"))
	s.Equal("admin", headers.Get("X-Locals-Role"))
}

func (s *GatewayClaimsMiddlewareSuite) TestMissingUserIDHeader_Returns401() {
	status, _ := s.do("/", map[string]string{"X-User-Role": "admin"})

	s.Equal(fiber.StatusUnauthorized, status)
}

func (s *GatewayClaimsMiddlewareSuite) TestMalformedUserIDHeader_Returns401() {
	status, _ := s.do("/", map[string]string{"X-User-Id": "not-a-uuid"})

	s.Equal(fiber.StatusUnauthorized, status)
}

func (s *GatewayClaimsMiddlewareSuite) TestMissingRoleHeader_StillSucceedsWithEmptyRole() {
	userID := uuid.New()
	status, headers := s.do("/", map[string]string{"X-User-Id": userID.String()})

	s.Equal(fiber.StatusOK, status)
	s.Equal(userID.String(), headers.Get("X-Locals-ID"))
	s.Empty(headers.Get("X-Locals-Role"))
}

func (s *GatewayClaimsMiddlewareSuite) TestAdminRole_PassesAdminOnly() {
	userID := uuid.New()
	status, _ := s.do("/admin", map[string]string{
		"X-User-Id":   userID.String(),
		"X-User-Role": "admin",
	})

	s.Equal(fiber.StatusOK, status)
}

func (s *GatewayClaimsMiddlewareSuite) TestNonAdminRole_BlockedByAdminOnly() {
	userID := uuid.New()
	status, _ := s.do("/admin", map[string]string{
		"X-User-Id":   userID.String(),
		"X-User-Role": "reader",
	})

	s.Equal(fiber.StatusForbidden, status)
}
