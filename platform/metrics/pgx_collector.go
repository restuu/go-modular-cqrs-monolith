package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// PgxCollector is a prometheus.Collector that reports pgxpool connection pool stats on every scrape.
// It reads pool.Stat() lazily — no background goroutine required.
type PgxCollector struct {
	pool          *pgxpool.Pool
	acquiredConns *prometheus.Desc
	idleConns     *prometheus.Desc
	totalConns    *prometheus.Desc
	maxConns      *prometheus.Desc
	acquireCount  *prometheus.Desc
}

// NewPgxCollector builds a collector for the given pool.
// The caller is responsible for registering it with a Prometheus registry.
func NewPgxCollector(pool *pgxpool.Pool) *PgxCollector {
	const ns = "pgxpool"
	return &PgxCollector{
		pool:          pool,
		acquiredConns: prometheus.NewDesc(ns+"_acquired_conns", "Number of currently acquired connections.", nil, nil),
		idleConns:     prometheus.NewDesc(ns+"_idle_conns", "Number of idle connections in the pool.", nil, nil),
		totalConns:    prometheus.NewDesc(ns+"_total_conns", "Total number of open connections managed by the pool.", nil, nil),
		maxConns:      prometheus.NewDesc(ns+"_max_conns", "Maximum number of connections allowed by the pool config.", nil, nil),
		acquireCount:  prometheus.NewDesc(ns+"_acquire_count_total", "Cumulative number of successful connection acquisitions.", nil, nil),
	}
}

// Describe sends the descriptors of each metric to the channel.
func (c *PgxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquiredConns
	ch <- c.idleConns
	ch <- c.totalConns
	ch <- c.maxConns
	ch <- c.acquireCount
}

// Collect reads pool stats and emits the current values.
func (c *PgxCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.pool.Stat()
	ch <- prometheus.MustNewConstMetric(c.acquiredConns, prometheus.GaugeValue, float64(s.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(s.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue, float64(s.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue, float64(s.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.acquireCount, prometheus.CounterValue, float64(s.AcquireCount()))
}
