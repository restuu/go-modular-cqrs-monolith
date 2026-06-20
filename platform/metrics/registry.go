// Package metrics provides Prometheus instrumentation for the API.
// Use New() to obtain a fresh registry, an HTTP recorder, and a pgxpool collector.
// Register the pgxpool collector after calling New() once a pgxpool is available.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// New creates an isolated Prometheus registry with the standard Go runtime and process
// collectors registered. It also registers and returns an HTTP Recorder.
// Using an explicit registry (not the global default) avoids pollution from third-party
// libraries that auto-register with prometheus.DefaultRegisterer.
func New() (*prometheus.Registry, *Recorder) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	rec := newRecorder(reg)
	return reg, rec
}
