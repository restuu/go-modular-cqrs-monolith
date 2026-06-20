package metrics_test

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-modular-cqrs-monolith/platform/metrics"
)

func gatherByName(t *testing.T, reg *prometheus.Registry, name string) *dto.MetricFamily {
	t.Helper()
	families, err := reg.Gather()
	require.NoError(t, err)
	for _, f := range families {
		if f.GetName() == name {
			return f
		}
	}
	return nil
}

func TestRecorder_ObserveRequest_increments_counter(t *testing.T) {
	reg, rec := metrics.New()
	rec.ObserveRequest("GET", "/api/v1/articles", 200, 10*time.Millisecond)
	rec.ObserveRequest("GET", "/api/v1/articles", 200, 20*time.Millisecond)
	rec.ObserveRequest("POST", "/api/v1/auth/login", 401, 5*time.Millisecond)

	family := gatherByName(t, reg, "http_requests_total")
	require.NotNil(t, family)

	totals := map[string]float64{}
	for _, m := range family.GetMetric() {
		labels := map[string]string{}
		for _, l := range m.GetLabel() {
			labels[l.GetName()] = l.GetValue()
		}
		key := labels["method"] + " " + labels["route"] + " " + labels["status"]
		totals[key] = m.GetCounter().GetValue()
	}

	assert.Equal(t, float64(2), totals["GET /api/v1/articles 200"])
	assert.Equal(t, float64(1), totals["POST /api/v1/auth/login 401"])
}

func TestRecorder_ObserveRequest_records_histogram(t *testing.T) {
	reg, rec := metrics.New()
	rec.ObserveRequest("GET", "/api/v1/articles", 200, 50*time.Millisecond)

	family := gatherByName(t, reg, "http_request_duration_seconds")
	require.NotNil(t, family)
	require.Len(t, family.GetMetric(), 1)

	h := family.GetMetric()[0].GetHistogram()
	assert.Equal(t, uint64(1), h.GetSampleCount())
	assert.InDelta(t, 0.05, h.GetSampleSum(), 0.001)

	// At least one bucket should have count > 0 (the 0.1s bucket covers 50ms)
	found := false
	for _, b := range h.GetBucket() {
		if b.GetCumulativeCount() > 0 {
			found = true
			break
		}
	}
	assert.True(t, found, "expected at least one non-empty histogram bucket")
}

func TestNew_registers_go_and_process_collectors(t *testing.T) {
	reg, _ := metrics.New()
	families, err := reg.Gather()
	require.NoError(t, err)

	names := make([]string, 0, len(families))
	for _, f := range families {
		names = append(names, f.GetName())
	}
	joined := strings.Join(names, ",")

	assert.Contains(t, joined, "go_goroutines", "go collector should be registered")
	assert.Contains(t, joined, "process_open_fds", "process collector should be registered")
}
