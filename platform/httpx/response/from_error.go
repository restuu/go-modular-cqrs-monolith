package response

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"go-modular-cqrs-monolith/platform/errcode"
)

// statusFor maps an errcode.Code to an HTTP status code.
// This is the single place where that mapping lives; the domain layer never touches HTTP.
func statusFor(code errcode.Code) int {
	switch code {
	case errcode.Unauthorized:
		return fiber.StatusUnauthorized
	case errcode.Forbidden:
		return fiber.StatusForbidden
	case errcode.ValidationError, errcode.InvalidBody, errcode.InvalidQuery,
		errcode.InvalidID, errcode.InvalidPagination, errcode.InvalidSourceID,
		errcode.MissingSlug, errcode.MissingTagSlug:
		return fiber.StatusBadRequest
	case errcode.InternalError:
		return fiber.StatusInternalServerError
	}
	s := string(code)
	switch {
	case strings.HasSuffix(s, "_NOT_FOUND"):
		return fiber.StatusNotFound
	case strings.HasSuffix(s, "_DUPLICATE"),
		strings.HasSuffix(s, "_TAKEN"),
		strings.HasPrefix(s, "ALREADY_"):
		return fiber.StatusConflict
	}
	return fiber.StatusBadRequest
}

// FromError writes the appropriate error response for err.
// If err is an *errcode.AppError the code and message are used directly.
// Any other error is logged at ERROR level and returns a generic 500.
func FromError(c fiber.Ctx, logger *slog.Logger, err error) error {
	if appErr, ok := errors.AsType[*errcode.AppError](err); ok {
		return c.Status(statusFor(appErr.Code)).JSON(Envelope{
			Success: false,
			Error:   &APIError{Code: string(appErr.Code), Message: appErr.Message},
		})
	}
	logger.ErrorContext(c.Context(), "unhandled error",
		slog.String("path", c.Path()),
		slog.String("method", c.Method()),
		slog.Any("error", err),
	)
	return c.Status(fiber.StatusInternalServerError).JSON(Envelope{
		Success: false,
		Error:   &APIError{Code: string(errcode.InternalError), Message: "an unexpected error occurred"},
	})
}
