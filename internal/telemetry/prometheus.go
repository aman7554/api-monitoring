package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pulsewatch_http_requests_total",
			Help: "Total number of HTTP requests processed by API",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pulsewatch_http_request_duration_seconds",
			Help:    "Histogram of response latency for HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	MonitorChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pulsewatch_monitor_checks_total",
			Help: "Total monitor check executions",
		},
		[]string{"type", "status"},
	)

	MonitorLatencyMS = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pulsewatch_monitor_latency_ms",
			Help:    "Latency of monitor targets in milliseconds",
			Buckets: []float64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		},
		[]string{"type", "status"},
	)

	ActiveIncidentsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "pulsewatch_active_incidents_count",
			Help: "Current total number of active ongoing incidents",
		},
	)
)
