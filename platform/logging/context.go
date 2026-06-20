// Package logging provides context-aware slog helpers for propagating
// request-scoped fields (currently: request_id) through the call stack.
package logging

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type contextKey struct{}

// WithRequestID stores a request ID in the context so it can be retrieved later.
func WithRequestID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// RequestIDFromContext returns the request ID stored by WithRequestID.
// Returns uuid.Nil and false when no ID is present.
func RequestIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(contextKey{}).(uuid.UUID)
	return id, ok
}

// FromContext returns base enriched with a request_id attribute when the
// context carries one (set by WithRequestID). Returns base unchanged otherwise.
func FromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if id, ok := RequestIDFromContext(ctx); ok {
		return base.With("request_id", id.String())
	}
	return base
}
