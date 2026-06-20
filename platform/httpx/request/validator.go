package request

import (
	"cmp"
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"go-modular-cqrs-monolith/platform/errcode"
	"go-modular-cqrs-monolith/platform/httpx/response"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	// Use json/query tag names in validation error messages instead of Go field names.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]; name != "" && name != "-" {
			return name
		}
		if name := fld.Tag.Get("query"); name != "" {
			return name
		}
		return fld.Name
	})
	return v
}

// ParseBody parses the JSON request body into dst and validates its struct tags.
// Returns the response error (already written) on failure; nil on success.
func ParseBody(c fiber.Ctx, dst any) error {
	if err := c.Bind().Body(dst); err != nil {
		return response.BadRequest(c, errcode.InvalidBody, "request body is not valid JSON")
	}
	return ValidateStruct(c, dst)
}

// ValidateStruct validates an already-populated struct against its validate tags.
// Returns the response error (already written) on failure; nil on success.
func ValidateStruct(c fiber.Ctx, dst any) error {
	if err := validate.Struct(dst); err != nil {
		var validationErrs validator.ValidationErrors
		if !errors.As(err, &validationErrs) {
			return response.BadRequest(c, errcode.ValidationError, err.Error())
		}
		details := make([]response.ValidationDetail, 0, len(validationErrs))
		for _, e := range validationErrs {
			details = append(details, response.ValidationDetail{
				Field:   e.Field(),
				Message: fieldErrMsg(e),
			})
		}
		return response.ValidationErrors(c, details)
	}
	return nil
}

type Pagination struct {
	Page    int `query:"page"`
	PerPage int `query:"per_page"`
}

// ParsePagination reads page and per_page from query params.
// Returns defaults (1, defaultPerPage) when params are absent or unparseable.
// Returns a 400 response error for explicitly provided out-of-range values.
func ParsePagination(c fiber.Ctx, defaultPerPage int) (page, perPage int, err error) {
	var p Pagination
	if err = c.Bind().Query(&p); err != nil {
		return 0, 0, response.BadRequest(c, errcode.InvalidPagination, "invalid pagination parameters")
	}

	page = cmp.Or(p.Page, 1)
	perPage = cmp.Or(p.PerPage, defaultPerPage)

	if c.Query("page") != "" && page < 1 {
		return 0, 0, response.BadRequest(c, errcode.InvalidPagination, "page must be at least 1")
	}
	if c.Query("per_page") != "" && (perPage < 1 || perPage > 100) {
		return 0, 0, response.BadRequest(c, errcode.InvalidPagination, "per_page must be between 1 and 100")
	}
	return page, perPage, nil
}

// staticMessages maps simple validator tags to fixed human-readable messages.
var staticMessages = map[string]string{
	"required": "is required",
	"email":    "must be a valid email address",
	"url":      "must be a valid URL",
	"alphanum": "must contain only alphanumeric characters",
	"dive":     "contains invalid items",
}

func fieldErrMsg(e validator.FieldError) string {
	if msg, ok := staticMessages[e.Tag()]; ok {
		return msg
	}
	switch e.Tag() {
	case "min":
		return minMsg(e.Kind(), e.Param())
	case "max":
		return maxMsg(e.Kind(), e.Param())
	case "oneof":
		return "must be one of: " + strings.ReplaceAll(e.Param(), " ", ", ")
	case "gte":
		return "must be " + e.Param() + " or greater"
	case "lte":
		return "must be " + e.Param() + " or less"
	default:
		return "failed validation (" + e.Tag() + ")"
	}
}

func minMsg(kind reflect.Kind, param string) string {
	switch kind {
	case reflect.String:
		return "must be at least " + param + " characters long"
	case reflect.Slice, reflect.Array:
		return "must have at least " + param + " items"
	default:
		return "must be at least " + param
	}
}

func maxMsg(kind reflect.Kind, param string) string {
	switch kind {
	case reflect.String:
		return "must be at most " + param + " characters long"
	case reflect.Slice, reflect.Array:
		return "must have at most " + param + " items"
	default:
		return "must be at most " + param
	}
}
