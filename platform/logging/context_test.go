package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"go-modular-cqrs-monolith/platform/logging"
)

type LoggingContextSuite struct {
	suite.Suite
}

func TestLoggingContextSuite(t *testing.T) {
	suite.Run(t, new(LoggingContextSuite))
}

func (s *LoggingContextSuite) TestWithRequestID_and_RequestIDFromContext() {
	s.Run("round_trip_returns_same_uuid", func() {
		id := uuid.New()
		ctx := logging.WithRequestID(context.Background(), id)

		got, ok := logging.RequestIDFromContext(ctx)

		s.True(ok)
		s.Equal(id, got)
	})

	s.Run("returns_false_when_context_has_no_id", func() {
		_, ok := logging.RequestIDFromContext(context.Background())
		s.False(ok)
	})

	s.Run("child_context_inherits_id", func() {
		id := uuid.New()
		parent := logging.WithRequestID(context.Background(), id)
		child := context.WithValue(parent, struct{ k string }{"extra"}, "v")

		got, ok := logging.RequestIDFromContext(child)

		s.True(ok)
		s.Equal(id, got)
	})
}

func (s *LoggingContextSuite) TestFromContext() {
	s.Run("attaches_request_id_attr_when_present", func() {
		id := uuid.New()
		ctx := logging.WithRequestID(context.Background(), id)

		buf := &bytes.Buffer{}
		base := slog.New(slog.NewJSONHandler(buf, nil))
		logger := logging.FromContext(ctx, base)
		logger.Info("test message")

		var record map[string]any
		s.Require().NoError(json.Unmarshal(buf.Bytes(), &record))
		s.Equal(id.String(), record["request_id"])
		s.Equal("test message", record["msg"])
	})

	s.Run("returns_base_unchanged_when_no_id_in_context", func() {
		buf := &bytes.Buffer{}
		base := slog.New(slog.NewJSONHandler(buf, nil))
		logger := logging.FromContext(context.Background(), base)
		logger.Info("no id")

		var record map[string]any
		s.Require().NoError(json.Unmarshal(buf.Bytes(), &record))
		_, hasID := record["request_id"]
		s.False(hasID)
	})
}
