package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

// PrometheusMetrics holds all Prometheus metrics for the MCP proxy
type PrometheusMetrics struct {
	// Counters
	RequestsTotal      *prometheus.CounterVec
	StatusCodesTotal   *prometheus.CounterVec
	LakeraBlocksTotal  *prometheus.CounterVec
	BackendErrorsTotal *prometheus.CounterVec
	RetriesTotal       *prometheus.CounterVec
	DLQMessagesTotal   *prometheus.CounterVec

	// Histograms
	RequestDuration *prometheus.HistogramVec

	// Gauges
	ActiveRequests      prometheus.Gauge
	DLQMessages         prometheus.Gauge
	CircuitBreakerState *prometheus.GaugeVec

	// Registry for the prometheus handler
	registry *prometheus.Registry
}

// NewPrometheusMetrics creates and registers all Prometheus metrics
func NewPrometheusMetrics() *PrometheusMetrics {
	registry := prometheus.NewRegistry()
	pm := &PrometheusMetrics{
		registry: registry,
	}

	// Register standard Go runtime metrics
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	// mcp_proxy_requests_total - Total requests with allowed status
	pm.RequestsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_proxy_requests_total",
			Help: "Total number of requests processed by the MCP proxy",
		},
		[]string{"method", "endpoint", "allowed"},
	)

	// mcp_proxy_status_codes_total - Total responses by status code
	pm.StatusCodesTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_proxy_status_codes_total",
			Help: "Total number of responses by HTTP status code",
		},
		[]string{"code"},
	)

	// mcp_proxy_lakera_blocks_total - Tools blocked by Lakera
	pm.LakeraBlocksTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_proxy_lakera_blocks_total",
			Help: "Total number of tools blocked by Lakera semantic firewall",
		},
		[]string{"reason"},
	)

	// mcp_proxy_backend_errors_total - Backend errors
	pm.BackendErrorsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_proxy_backend_errors_total",
			Help: "Total number of backend errors by type",
		},
		[]string{"error_type"},
	)

	// mcp_proxy_retries_total - Retry attempts
	pm.RetriesTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_proxy_retries_total",
			Help: "Total number of retry attempts",
		},
		[]string{"endpoint"},
	)

	// mcp_proxy_dlq_messages_total - Messages in Dead Letter Queue
	pm.DLQMessagesTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_proxy_dlq_messages_total",
			Help: "Total number of messages sent to DLQ",
		},
		[]string{"source"},
	)

	// mcp_proxy_request_duration_seconds - Request latency histogram
	pm.RequestDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_proxy_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// mcp_proxy_active_requests - Currently active requests
	pm.ActiveRequests = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_proxy_active_requests",
			Help: "Number of requests currently being processed",
		},
	)

	// mcp_proxy_dlq_messages - Current DLQ message count
	pm.DLQMessages = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_proxy_dlq_messages",
			Help: "Current number of messages in the Dead Letter Queue",
		},
	)

	// mcp_proxy_circuit_breaker_state - Circuit breaker state (0=closed, 1=open, 2=half-open)
	pm.CircuitBreakerState = promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_proxy_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"endpoint"},
	)

	return pm
}

// RecordRequest records a request in Prometheus metrics
func (pm *PrometheusMetrics) RecordRequest(method, endpoint string, allowed bool, statusCode int, duration time.Duration) {
	allowedStr := "false"
	if allowed {
		allowedStr = "true"
	}

	pm.RequestsTotal.WithLabelValues(method, endpoint, allowedStr).Inc()
	pm.StatusCodesTotal.WithLabelValues(strconv.Itoa(statusCode)).Inc()
	pm.RequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordLakeraBlock records a Lakera block
func (pm *PrometheusMetrics) RecordLakeraBlock(reason string) {
	pm.LakeraBlocksTotal.WithLabelValues(reason).Inc()
}

// RecordBackendError records a backend error
func (pm *PrometheusMetrics) RecordBackendError(errorType string) {
	pm.BackendErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordRetry records a retry attempt
func (pm *PrometheusMetrics) RecordRetry(endpoint string) {
	pm.RetriesTotal.WithLabelValues(endpoint).Inc()
}

// RecordDLQMessage records a message sent to DLQ
func (pm *PrometheusMetrics) RecordDLQMessage(source string) {
	pm.DLQMessagesTotal.WithLabelValues(source).Inc()
}

// SetDLQCount sets the current DLQ message count
func (pm *PrometheusMetrics) SetDLQCount(count int) {
	pm.DLQMessages.Set(float64(count))
}

// SetCircuitBreakerState sets the circuit breaker state for an endpoint
func (pm *PrometheusMetrics) SetCircuitBreakerState(endpoint string, state CircuitState) {
	pm.CircuitBreakerState.WithLabelValues(endpoint).Set(float64(state))
}

// IncActiveRequests increments the active requests gauge
func (pm *PrometheusMetrics) IncActiveRequests() {
	pm.ActiveRequests.Inc()
}

// DecActiveRequests decrements the active requests gauge
func (pm *PrometheusMetrics) DecActiveRequests() {
	pm.ActiveRequests.Dec()
}

// Gather implements prometheus.Gatherer interface
func (pm *PrometheusMetrics) Gather() ([]*dto.MetricFamily, error) {
	return pm.registry.Gather()
}

// NewPrometheusHandler creates an HTTP handler for the /metrics endpoint
// Returns a handler that exposes metrics in Prometheus text format 0.0.4
func NewPrometheusHandler(pm *PrometheusMetrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Type for Prometheus text format 0.0.4
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		// Use the standard promhttp handler for correct formatting
		handler := promhttp.HandlerFor(pm, promhttp.HandlerOpts{
			EnableOpenMetrics: false,
		})
		handler.ServeHTTP(w, r)
	}
}
