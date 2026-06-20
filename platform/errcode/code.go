package errcode

// Code is a machine-readable string identifier for an error condition.
// Cross-cutting codes live here; domain-specific codes live next to their domain package.
type Code string

const (
	// 401
	Unauthorized Code = "UNAUTHORIZED"

	// 403
	Forbidden Code = "FORBIDDEN"

	// 500
	InternalError Code = "INTERNAL_ERROR"

	// 400 — request/validation errors (used by validator and request parsing helpers)
	ValidationError   Code = "VALIDATION_ERROR"
	InvalidBody       Code = "INVALID_BODY"
	InvalidQuery      Code = "INVALID_QUERY"
	InvalidID         Code = "INVALID_ID"
	InvalidPagination Code = "INVALID_PAGINATION"
	InvalidSourceID   Code = "INVALID_SOURCE_ID"
	MissingSlug       Code = "MISSING_SLUG"
	MissingTagSlug    Code = "MISSING_TAG_SLUG"
)
