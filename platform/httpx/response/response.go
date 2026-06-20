package response

import (
	"github.com/gofiber/fiber/v3"

	"go-modular-cqrs-monolith/platform/errcode"
)

// Envelope is the standard API response wrapper.
type Envelope struct {
	Success bool      `json:"success"`
	Data    any       `json:"data"`
	Error   *APIError `json:"error"`
	Meta    any       `json:"meta"`
}

// APIError carries a machine-readable code and human-readable message.
// Details is populated only for validation errors.
type APIError struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []ValidationDetail `json:"details,omitempty"`
}

// ValidationDetail describes a single field validation failure.
type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// PaginationMeta carries pagination info for list responses.
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func OK(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Success: true, Data: data})
}

func OKWithMeta(c fiber.Ctx, data any, meta any) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Success: true, Data: data, Meta: meta})
}

func Created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Success: true, Data: data})
}

func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func BadRequest(c fiber.Ctx, code errcode.Code, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(Envelope{
		Success: false, Error: &APIError{Code: string(code), Message: message},
	})
}

func Unauthorized(c fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(Envelope{
		Success: false, Error: &APIError{Code: string(errcode.Unauthorized), Message: "authentication required"},
	})
}

func Forbidden(c fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(Envelope{
		Success: false, Error: &APIError{Code: string(errcode.Forbidden), Message: "insufficient permissions"},
	})
}

func NotFound(c fiber.Ctx, code errcode.Code, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(Envelope{
		Success: false, Error: &APIError{Code: string(code), Message: message},
	})
}

func Conflict(c fiber.Ctx, code errcode.Code, message string) error {
	return c.Status(fiber.StatusConflict).JSON(Envelope{
		Success: false, Error: &APIError{Code: string(code), Message: message},
	})
}

func InternalServerError(c fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(Envelope{
		Success: false, Error: &APIError{Code: string(errcode.InternalError), Message: "an unexpected error occurred"},
	})
}

func ValidationErrors(c fiber.Ctx, details []ValidationDetail) error {
	return c.Status(fiber.StatusBadRequest).JSON(Envelope{
		Success: false,
		Error: &APIError{
			Code:    string(errcode.ValidationError),
			Message: "request validation failed",
			Details: details,
		},
	})
}
