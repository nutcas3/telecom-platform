package main

import (
	"context"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/metrics"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// startMetricsCollection periodically collects and stores metrics.
func startMetricsCollection(ctx context.Context, mc *metrics.MetricsCollectorWithTimeSeries, db *database.Database) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			collectMetrics(ctx, mc, db)
		}
	}
}

// collectMetrics gathers metrics from various sources.
func collectMetrics(ctx context.Context, mc *metrics.MetricsCollectorWithTimeSeries, db *database.Database) {
	totalSubs, activeSubs, suspendedSubs := getSubscriberStats(db)
	mc.UpdateSubscriberMetrics(totalSubs, activeSubs, suspendedSubs)

	systemMetrics := metrics.SystemMetrics{
		ActiveSessions: getActiveSessionsCount(db),
		CPUUsage:       getCPUUsage(),
		MemoryUsage:    getMemoryUsage(),
		NetworkRX:      getNetworkRX(),
		NetworkTX:      getNetworkTX(),
		Timestamp:      time.Now(),
	}

	if ts := mc.GetTimeSeriesStorage(); ts != nil {
		if err := ts.StoreSystemMetrics(ctx, systemMetrics); err != nil {
			log.Printf("Failed to store system metrics: %v", err)
		}
	}
}

// getSubscriberStats counts subscribers by status.
func getSubscriberStats(db *database.Database) (total, active, suspended int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var totalCount, activeCount, suspendedCount int64
	if err := db.DB.WithContext(ctx).Model(&models.Subscriber{}).Count(&totalCount).Error; err != nil {
		log.Printf("Error counting total subscribers: %v", err)
		return 0, 0, 0
	}
	if err := db.DB.WithContext(ctx).Model(&models.Subscriber{}).Where("status = ?", "active").Count(&activeCount).Error; err != nil {
		log.Printf("Error counting active subscribers: %v", err)
		return int(totalCount), 0, 0
	}
	if err := db.DB.WithContext(ctx).Model(&models.Subscriber{}).
		Where("status IN ?", []string{"suspended", "deactivated", "blocked"}).
		Count(&suspendedCount).Error; err != nil {
		log.Printf("Error counting suspended subscribers: %v", err)
		return int(totalCount), int(activeCount), 0
	}
	return int(totalCount), int(activeCount), int(suspendedCount)
}

// getActiveSessionsCount returns the number of currently active charging sessions.
func getActiveSessionsCount(db *database.Database) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	if err := db.DB.WithContext(ctx).Model(&models.Session{}).Where(
		"status IN ? AND end_time IS NULL",
		[]string{"active", "connected", "established"},
	).Count(&count).Error; err != nil {
		log.Printf("Error counting active sessions: %v", err)
		return 0
	}

	return int(count)
}

// getCPUUsage reads /proc/stat and returns the non-idle CPU percentage.
func getCPUUsage() float64 {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		log.Printf("Error reading /proc/stat: %v", err)
		return 0
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0
	}

	fields := strings.Fields(lines[0])
	if len(fields) < 8 || fields[0] != "cpu" {
		return 0
	}

	var totalCPU, idleCPU float64
	for i, field := range fields[1:8] {
		val, err := strconv.ParseFloat(field, 64)
		if err != nil {
			continue
		}
		totalCPU += val
		if i == 3 {
			idleCPU = val
		}
	}

	if totalCPU == 0 {
		return 0
	}

	return math.Min(((totalCPU-idleCPU)/totalCPU)*100, 100.0)
}

// getMemoryUsage reads /proc/meminfo and returns used-memory percentage.
func getMemoryUsage() float64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		log.Printf("Error reading /proc/meminfo: %v", err)
		return 0
	}

	lines := strings.Split(string(data), "\n")
	var totalMem, availableMem float64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				totalMem = val * 1024
			}
		case "MemAvailable:":
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				availableMem = val * 1024
			}
		}
	}

	if totalMem == 0 {
		return 0
	}

	return math.Min(((totalMem-availableMem)/totalMem)*100, 100.0)
}

// getNetworkRX returns total RX bytes across non-loopback interfaces.
func getNetworkRX() int64 { return readNetworkStat(1) }

// getNetworkTX returns total TX bytes across non-loopback interfaces.
func getNetworkTX() int64 { return readNetworkStat(9) }

// readNetworkStat parses /proc/net/dev for a given field index.
func readNetworkStat(fieldIdx int) int64 {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		log.Printf("Error reading /proc/net/dev: %v", err)
		return 0
	}

	lines := strings.Split(string(data), "\n")
	var total int64
	for i, line := range lines {
		if i < 2 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		if strings.HasPrefix(fields[0], "lo:") {
			continue
		}
		if v, err := strconv.ParseInt(fields[fieldIdx], 10, 64); err == nil {
			total += v
		}
	}
	return total
}
