package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telecom_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "telecom_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Subscriber metrics
	subscribersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "telecom_subscribers_total",
			Help: "Total number of subscribers",
		},
	)

	subscribersActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "telecom_subscribers_active",
			Help: "Number of active subscribers",
		},
	)

	subscribersSuspended = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "telecom_subscribers_suspended",
			Help: "Number of suspended subscribers",
		},
	)

	// Usage metrics
	dataUsageBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telecom_data_usage_bytes_total",
			Help: "Total data usage in bytes",
		},
		[]string{"subscriber_id", "direction"},
	)

	voiceUsageSeconds = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telecom_voice_usage_seconds_total",
			Help: "Total voice usage in seconds",
		},
		[]string{"subscriber_id"},
	)

	smsUsageTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telecom_sms_usage_total",
			Help: "Total SMS count",
		},
		[]string{"subscriber_id"},
	)

	// Charging metrics
	creditBalance = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "telecom_credit_balance",
			Help: "Subscriber credit balance",
		},
		[]string{"subscriber_id"},
	)

	paymentTransactionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telecom_payment_transactions_total",
			Help: "Total payment transactions",
		},
		[]string{"status", "gateway"},
	)

	// System metrics
	activeSessions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "telecom_active_sessions",
			Help: "Number of active sessions",
		},
	)

	systemUptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "telecom_system_uptime_seconds",
			Help: "System uptime in seconds",
		},
	)

	// Database metrics
	dbConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "telecom_db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "telecom_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Redis metrics
	redisConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "telecom_redis_connections_active",
			Help: "Number of active Redis connections",
		},
	)

	redisOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telecom_redis_operations_total",
			Help: "Total Redis operations",
		},
		[]string{"operation", "status"},
	)

	// Chaos engineering metrics
	chaosExperimentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telecom_chaos_experiments_total",
			Help: "Total chaos experiments run",
		},
		[]string{"type", "status"},
	)

	// eBPF packet gateway metrics
	packetsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telecom_packets_processed_total",
			Help: "Total packets processed by eBPF gateway",
		},
		[]string{"action"},
	)

	packetsDropped = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telecom_packets_dropped_total",
			Help: "Total packets dropped by eBPF gateway",
		},
		[]string{"reason"},
	)
)

// MetricsCollector handles Prometheus metrics collection
type MetricsCollector struct {
	startTime time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
	}
}

// Register registers all metrics with Prometheus
func (mc *MetricsCollector) Register() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(subscribersTotal)
	prometheus.MustRegister(subscribersActive)
	prometheus.MustRegister(subscribersSuspended)
	prometheus.MustRegister(dataUsageBytes)
	prometheus.MustRegister(voiceUsageSeconds)
	prometheus.MustRegister(smsUsageTotal)
	prometheus.MustRegister(creditBalance)
	prometheus.MustRegister(paymentTransactionsTotal)
	prometheus.MustRegister(activeSessions)
	prometheus.MustRegister(systemUptime)
	prometheus.MustRegister(dbConnectionsActive)
	prometheus.MustRegister(dbQueryDuration)
	prometheus.MustRegister(redisConnectionsActive)
	prometheus.MustRegister(redisOperationsTotal)
	prometheus.MustRegister(chaosExperimentsTotal)
	prometheus.MustRegister(packetsProcessed)
	prometheus.MustRegister(packetsDropped)

	// Start system uptime goroutine
	go mc.updateSystemMetrics()
}

// updateSystemMetrics updates system-level metrics periodically
func (mc *MetricsCollector) updateSystemMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		systemUptime.Set(time.Since(mc.startTime).Seconds())
	}
}

// HTTPMiddleware returns middleware to track HTTP metrics
func (mc *MetricsCollector) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Process request
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := rw.statusCode

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, http.StatusText(statusCode)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// responseWriter is a wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// UpdateSubscriberMetrics updates subscriber-related metrics
func (mc *MetricsCollector) UpdateSubscriberMetrics(total, active, suspended int) {
	subscribersTotal.Set(float64(total))
	subscribersActive.Set(float64(active))
	subscribersSuspended.Set(float64(suspended))
}

// RecordDataUsage records data usage for a subscriber
func (mc *MetricsCollector) RecordDataUsage(subscriberID string, direction string, bytes int64) {
	dataUsageBytes.WithLabelValues(subscriberID, direction).Add(float64(bytes))
}

// RecordVoiceUsage records voice usage for a subscriber
func (mc *MetricsCollector) RecordVoiceUsage(subscriberID string, seconds int64) {
	voiceUsageSeconds.WithLabelValues(subscriberID).Add(float64(seconds))
}

// RecordSMSUsage records SMS usage for a subscriber
func (mc *MetricsCollector) RecordSMSUsage(subscriberID string, count int64) {
	smsUsageTotal.WithLabelValues(subscriberID).Add(float64(count))
}

// UpdateCreditBalance updates credit balance for a subscriber
func (mc *MetricsCollector) UpdateCreditBalance(subscriberID string, balance float64) {
	creditBalance.WithLabelValues(subscriberID).Set(balance)
}

// RecordPaymentTransaction records a payment transaction
func (mc *MetricsCollector) RecordPaymentTransaction(status, gateway string) {
	paymentTransactionsTotal.WithLabelValues(status, gateway).Inc()
}

// UpdateActiveSessions updates the active sessions count
func (mc *MetricsCollector) UpdateActiveSessions(count int) {
	activeSessions.Set(float64(count))
}

// UpdateDBMetrics updates database metrics
func (mc *MetricsCollector) UpdateDBMetrics(activeConnections int, operation string, duration time.Duration) {
	dbConnectionsActive.Set(float64(activeConnections))
	dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// UpdateRedisMetrics updates Redis metrics
func (mc *MetricsCollector) UpdateRedisMetrics(activeConnections int, operation, status string) {
	redisConnectionsActive.Set(float64(activeConnections))
	redisOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordChaosExperiment records a chaos experiment
func (mc *MetricsCollector) RecordChaosExperiment(experimentType, status string) {
	chaosExperimentsTotal.WithLabelValues(experimentType, status).Inc()
}

// RecordPacketMetrics records packet processing metrics
func (mc *MetricsCollector) RecordPacketMetrics(action string, count int64, reason string) {
	if action == "dropped" {
		packetsDropped.WithLabelValues(reason).Add(float64(count))
	} else {
		packetsProcessed.WithLabelValues(action).Add(float64(count))
	}
}

// Handler returns the Prometheus metrics HTTP handler
func (mc *MetricsCollector) Handler() http.Handler {
	return promhttp.Handler()
}

// StartMetricsServer starts the Prometheus metrics server
func (mc *MetricsCollector) StartMetricsServer(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", mc.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}
