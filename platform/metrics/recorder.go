package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Recorder holds the HTTP request metrics registered with a Prometheus registry.
type Recorder struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func newRecorder(reg prometheus.Registerer) *Recorder {
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests, partitioned by method, route, and status code.",
	}, []string{"method", "route", "status"})

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency distribution, partitioned by method, route, and status code.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "status"})

	reg.MustRegister(requests, duration)
	return &Recorder{requests: requests, duration: duration}
}

// ObserveRequest records one completed HTTP request.
func (r *Recorder) ObserveRequest(method, route string, statusCode int, dur time.Duration) {
	status := strconv.Itoa(statusCode)
	r.requests.WithLabelValues(method, route, status).Inc()
	r.duration.WithLabelValues(method, route, status).Observe(dur.Seconds())
}
