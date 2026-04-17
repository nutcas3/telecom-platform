package metrics

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// TimeSeriesStorage handles advanced time-series metrics storage
type TimeSeriesStorage struct {
	client   api.WriteAPIBlocking
	queryAPI api.QueryAPI
	org      string
	bucket   string
}

// MetricPoint represents a single metric data point
type MetricPoint struct {
	Measurement string            `json:"measurement"`
	Tags        map[string]string `json:"tags"`
	Fields      map[string]any    `json:"fields"`
	Timestamp   time.Time         `json:"timestamp"`
}

// UsageMetrics represents usage metrics for time-series storage
type UsageMetrics struct {
	SubscriberID string    `json:"subscriber_id"`
	IMSI         string    `json:"imsi"`
	DataUp       int64     `json:"data_up"`
	DataDown     int64     `json:"data_down"`
	VoiceSeconds int64     `json:"voice_seconds"`
	SMSCount     int64     `json:"sms_count"`
	Timestamp    time.Time `json:"timestamp"`
}

// SystemMetrics represents system metrics for time-series storage
type SystemMetrics struct {
	ActiveSessions int       `json:"active_sessions"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    float64   `json:"memory_usage"`
	NetworkRX      int64     `json:"network_rx"`
	NetworkTX      int64     `json:"network_tx"`
	Timestamp      time.Time `json:"timestamp"`
}

// ChargingMetrics represents charging metrics for time-series storage
type ChargingMetrics struct {
	SubscriberID  string    `json:"subscriber_id"`
	Balance       float64   `json:"balance"`
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewTimeSeriesStorage creates a new time-series storage client
func NewTimeSeriesStorage(url, token, org, bucket string) (*TimeSeriesStorage, error) {
	client := influxdb2.NewClient(url, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)
	queryAPI := client.QueryAPI(org)

	return &TimeSeriesStorage{
		client:   writeAPI,
		queryAPI: queryAPI,
		org:      org,
		bucket:   bucket,
	}, nil
}

// StoreUsageMetrics stores usage metrics in InfluxDB
func (ts *TimeSeriesStorage) StoreUsageMetrics(ctx context.Context, metrics UsageMetrics) error {
	point := influxdb2.NewPoint(
		"usage",
		map[string]string{
			"subscriber_id": metrics.SubscriberID,
			"imsi":          metrics.IMSI,
		},
		map[string]interface{}{
			"data_up":       metrics.DataUp,
			"data_down":     metrics.DataDown,
			"voice_seconds": metrics.VoiceSeconds,
			"sms_count":     metrics.SMSCount,
		},
		metrics.Timestamp,
	)

	return ts.client.WritePoint(ctx, point)
}

// StoreSystemMetrics stores system metrics in InfluxDB
func (ts *TimeSeriesStorage) StoreSystemMetrics(ctx context.Context, metrics SystemMetrics) error {
	point := influxdb2.NewPoint(
		"system",
		map[string]string{},
		map[string]interface{}{
			"active_sessions": metrics.ActiveSessions,
			"cpu_usage":       metrics.CPUUsage,
			"memory_usage":    metrics.MemoryUsage,
			"network_rx":      metrics.NetworkRX,
			"network_tx":      metrics.NetworkTX,
		},
		metrics.Timestamp,
	)

	return ts.client.WritePoint(ctx, point)
}

// StoreChargingMetrics stores charging metrics in InfluxDB
func (ts *TimeSeriesStorage) StoreChargingMetrics(ctx context.Context, metrics ChargingMetrics) error {
	point := influxdb2.NewPoint(
		"charging",
		map[string]string{
			"subscriber_id":  metrics.SubscriberID,
			"transaction_id": metrics.TransactionID,
			"currency":       metrics.Currency,
			"status":         metrics.Status,
		},
		map[string]interface{}{
			"balance": metrics.Balance,
			"amount":  metrics.Amount,
		},
		metrics.Timestamp,
	)

	return ts.client.WritePoint(ctx, point)
}

// StorePacketMetrics stores packet processing metrics in InfluxDB
func (ts *TimeSeriesStorage) StorePacketMetrics(ctx context.Context, subscriberID string, packetsProcessed, packetsDropped int64, bytesProcessed int64) error {
	point := influxdb2.NewPoint(
		"packets",
		map[string]string{
			"subscriber_id": subscriberID,
		},
		map[string]interface{}{
			"packets_processed": packetsProcessed,
			"packets_dropped":   packetsDropped,
			"bytes_processed":   bytesProcessed,
		},
		time.Now(),
	)

	return ts.client.WritePoint(ctx, point)
}

// GetUsageStats retrieves usage statistics for a time range
func (ts *TimeSeriesStorage) GetUsageStats(ctx context.Context, subscriberID string, start, end time.Time) (*UsageMetrics, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r._measurement == "usage")
		|> filter(fn: (r) => r.subscriber_id == "%s")
		|> last()
	`, ts.bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), subscriberID)

	result, err := ts.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage stats: %w", err)
	}

	var metrics UsageMetrics
	for result.Next() {
		record := result.Record()
		metrics = UsageMetrics{
			SubscriberID: record.ValueByKey("subscriber_id").(string),
			IMSI:         record.ValueByKey("imsi").(string),
			DataUp:       record.ValueByKey("data_up").(int64),
			DataDown:     record.ValueByKey("data_down").(int64),
			VoiceSeconds: record.ValueByKey("voice_seconds").(int64),
			SMSCount:     record.ValueByKey("sms_count").(int64),
			Timestamp:    record.Time(),
		}
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return &metrics, nil
}

// GetSystemMetricsHistory retrieves system metrics history
func (ts *TimeSeriesStorage) GetSystemMetricsHistory(ctx context.Context, start, end time.Time, interval string) ([]SystemMetrics, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r._measurement == "system")
		|> aggregateWindow(every: %s, fn: mean, createEmpty: false)
		|> yield(name: "mean")
	`, ts.bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), interval)

	result, err := ts.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query system metrics: %w", err)
	}

	var metrics []SystemMetrics
	for result.Next() {
		record := result.Record()
		metric := SystemMetrics{
			ActiveSessions: int(record.ValueByKey("active_sessions").(int64)),
			CPUUsage:       record.ValueByKey("cpu_usage").(float64),
			MemoryUsage:    record.ValueByKey("memory_usage").(float64),
			NetworkRX:      record.ValueByKey("network_rx").(int64),
			NetworkTX:      record.ValueByKey("network_tx").(int64),
			Timestamp:      record.Time(),
		}
		metrics = append(metrics, metric)
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return metrics, nil
}

// GetSubscriberUsageHistory retrieves usage history for a subscriber
func (ts *TimeSeriesStorage) GetSubscriberUsageHistory(ctx context.Context, subscriberID string, start, end time.Time, interval string) ([]UsageMetrics, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r._measurement == "usage")
		|> filter(fn: (r) => r.subscriber_id == "%s")
		|> aggregateWindow(every: %s, fn: sum, createEmpty: false)
		|> yield(name: "sum")
	`, ts.bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), subscriberID, interval)

	result, err := ts.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriber usage: %w", err)
	}

	var metrics []UsageMetrics
	for result.Next() {
		record := result.Record()
		metric := UsageMetrics{
			SubscriberID: subscriberID,
			IMSI:         record.ValueByKey("imsi").(string),
			DataUp:       record.ValueByKey("data_up").(int64),
			DataDown:     record.ValueByKey("data_down").(int64),
			VoiceSeconds: record.ValueByKey("voice_seconds").(int64),
			SMSCount:     record.ValueByKey("sms_count").(int64),
			Timestamp:    record.Time(),
		}
		metrics = append(metrics, metric)
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return metrics, nil
}

// GetChargingHistory retrieves charging history for a subscriber
func (ts *TimeSeriesStorage) GetChargingHistory(ctx context.Context, subscriberID string, start, end time.Time) ([]ChargingMetrics, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r._measurement == "charging")
		|> filter(fn: (r) => r.subscriber_id == "%s")
		|> sort(columns: ["_time"])
	`, ts.bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), subscriberID)

	result, err := ts.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query charging history: %w", err)
	}

	var metrics []ChargingMetrics
	for result.Next() {
		record := result.Record()
		metric := ChargingMetrics{
			SubscriberID:  record.ValueByKey("subscriber_id").(string),
			TransactionID: record.ValueByKey("transaction_id").(string),
			Balance:       record.ValueByKey("balance").(float64),
			Amount:        record.ValueByKey("amount").(float64),
			Currency:      record.ValueByKey("currency").(string),
			Status:        record.ValueByKey("status").(string),
			Timestamp:     record.Time(),
		}
		metrics = append(metrics, metric)
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return metrics, nil
}

// GetTopDataUsers retrieves top data users for a time period
func (ts *TimeSeriesStorage) GetTopDataUsers(ctx context.Context, start, end time.Time, limit int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r._measurement == "usage")
		|> group(columns: ["subscriber_id"])
		|> sum(columns: ["data_up", "data_down"])
		|> map(fn: (r) => ({r with total_data: r.data_up + r.data_down}))
		|> sort(columns: ["total_data"], desc: true)
		|> limit(n: %d)
	`, ts.bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), limit)

	result, err := ts.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query top data users: %w", err)
	}

	var users []map[string]interface{}
	for result.Next() {
		record := result.Record()
		user := map[string]interface{}{
			"subscriber_id": record.ValueByKey("subscriber_id"),
			"total_data":    record.ValueByKey("total_data"),
			"data_up":       record.ValueByKey("data_up"),
			"data_down":     record.ValueByKey("data_down"),
		}
		users = append(users, user)
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return users, nil
}

// Close closes the InfluxDB client connection
func (ts *TimeSeriesStorage) Close() {
	// InfluxDB client doesn't need explicit closing in v2
}

// MetricsCollectorWithTimeSeries extends MetricsCollector with time-series storage
type MetricsCollectorWithTimeSeries struct {
	*MetricsCollector
	timeSeries *TimeSeriesStorage
}

// NewMetricsCollectorWithTimeSeries creates a new metrics collector with time-series storage
func NewMetricsCollectorWithTimeSeries(prometheusCollector *MetricsCollector, timeSeries *TimeSeriesStorage) *MetricsCollectorWithTimeSeries {
	return &MetricsCollectorWithTimeSeries{
		MetricsCollector: prometheusCollector,
		timeSeries:       timeSeries,
	}
}

// StoreUsageMetricsInTimeSeries stores usage metrics in both Prometheus and time-series
func (mc *MetricsCollectorWithTimeSeries) StoreUsageMetricsInTimeSeries(ctx context.Context, metrics UsageMetrics) error {
	// Update Prometheus metrics
	mc.RecordDataUsage(metrics.SubscriberID, "up", metrics.DataUp)
	mc.RecordDataUsage(metrics.SubscriberID, "down", metrics.DataDown)
	mc.RecordVoiceUsage(metrics.SubscriberID, metrics.VoiceSeconds)
	mc.RecordSMSUsage(metrics.SubscriberID, metrics.SMSCount)

	// Store in time-series
	return mc.timeSeries.StoreUsageMetrics(ctx, metrics)
}

// StoreSystemMetricsInTimeSeries stores system metrics in both Prometheus and time-series
func (mc *MetricsCollectorWithTimeSeries) StoreSystemMetricsInTimeSeries(ctx context.Context, metrics SystemMetrics) error {
	// Update Prometheus metrics
	mc.UpdateActiveSessions(metrics.ActiveSessions)

	// Store in time-series
	return mc.timeSeries.StoreSystemMetrics(ctx, metrics)
}

// StoreChargingMetricsInTimeSeries stores charging metrics in both Prometheus and time-series
func (mc *MetricsCollectorWithTimeSeries) StoreChargingMetricsInTimeSeries(ctx context.Context, metrics ChargingMetrics) error {
	// Update Prometheus metrics
	mc.UpdateCreditBalance(metrics.SubscriberID, metrics.Balance)
	mc.RecordPaymentTransaction(metrics.Status, "stripe")

	// Store in time-series
	return mc.timeSeries.StoreChargingMetrics(ctx, metrics)
}

// GetTimeSeriesStorage returns the time-series storage instance
func (mc *MetricsCollectorWithTimeSeries) GetTimeSeriesStorage() *TimeSeriesStorage {
	return mc.timeSeries
}
