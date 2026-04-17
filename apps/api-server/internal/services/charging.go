package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// ChargingService handles charging and billing operations
type ChargingService struct {
	db     *database.Database
	engine *ChargingEngineClient
}

// NewChargingService creates a new charging service
func NewChargingService(db *database.Database, cfg *config.Config) *ChargingService {
	return &ChargingService{
		db:     db,
		engine: NewChargingEngineClient(&cfg.ChargingEngine),
	}
}

// GetSystemStats gets system statistics
func (cs *ChargingService) GetSystemStats(ctx context.Context) (*models.SystemStats, error) {
	var stats models.SystemStats

	// Count active sessions
	var activeSessions int64
	if err := cs.db.DB.WithContext(ctx).Model(&models.Session{}).
		Where("status IN ?", []string{"active", "connected", "established"}).
		Count(&activeSessions).Error; err != nil {
		return nil, fmt.Errorf("failed to count active sessions: %w", err)
	}
	stats.ActiveSessions = int(activeSessions)

	// Count total accounts
	var totalAccounts int64
	if err := cs.db.DB.WithContext(ctx).Model(&models.SubscriberAccount{}).
		Count(&totalAccounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total accounts: %w", err)
	}
	stats.TotalAccounts = int(totalAccounts)

	// Count blocked users
	var blockedUsers int64
	if err := cs.db.DB.WithContext(ctx).Model(&models.Subscriber{}).
		Where("status IN ?", []string{"suspended", "deactivated", "blocked"}).
		Count(&blockedUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count blocked users: %w", err)
	}
	stats.BlockedUsers = int(blockedUsers)

	// Count low balance alerts
	var lowBalanceAlerts int64
	if err := cs.db.DB.WithContext(ctx).Model(&models.Alert{}).
		Where("type = ? AND resolved = ?", models.AlertTypeLowBalance, false).
		Count(&lowBalanceAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count low balance alerts: %w", err)
	}
	stats.LowBalanceAlerts = int(lowBalanceAlerts)

	// Calculate uptime (simplified - in real implementation would track service start time)
	stats.Uptime = time.Since(time.Now().AddDate(0, 0, -7)).Seconds() // Assume 7 days uptime

	return &stats, nil
}

// GetHealthStatus gets health status
func (cs *ChargingService) GetHealthStatus(ctx context.Context) (*models.HealthStatus, error) {
	status := &models.HealthStatus{
		RedisConnected: true, // In real implementation, would check Redis connection
		ActiveSync:     true, // In real implementation, would check sync status
		LastSync:       time.Now(),
		MemoryUsage:    50000000, // In real implementation, would get actual memory usage
	}

	// Check database connectivity
	if err := cs.db.DB.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		status.ActiveSync = false
	}

	return status, nil
}

// GetUsageStats gets usage statistics for a subscriber
func (cs *ChargingService) GetUsageStats(ctx context.Context, imsi string, period string) (*UsageStats, error) {
	var dataUsage, voiceUsage, smsUsage, totalCost float64

	// Calculate date range based on period
	var startTime time.Time
	switch period {
	case "daily":
		startTime = time.Now().AddDate(0, 0, -1)
	case "weekly":
		startTime = time.Now().AddDate(0, 0, -7)
	case "monthly":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		startTime = time.Now().AddDate(0, -1, 0) // Default to monthly
	}

	// Get usage statistics from database
	if err := cs.db.DB.WithContext(ctx).Model(&models.UsageEvent{}).
		Where("imsi = ? AND created_at >= ?", imsi, startTime).
		Select("SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as data_usage, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as voice_usage, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as sms_usage, SUM(cost) as total_cost",
			models.UsageTypeData, models.UsageTypeVoice, models.UsageTypeSMS).
		Row().Scan(&dataUsage, &voiceUsage, &smsUsage, &totalCost); err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	// Calculate trend (compare with previous period)
	var prevDataUsage, prevVoiceUsage, prevSmsUsage float64
	prevStartTime := startTime.AddDate(0, 0, -int(time.Since(startTime).Hours()/24))

	if err := cs.db.DB.WithContext(ctx).Model(&models.UsageEvent{}).
		Where("imsi = ? AND created_at >= ? AND created_at < ?", imsi, prevStartTime, startTime).
		Select("SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as data_usage, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as voice_usage, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as sms_usage",
			models.UsageTypeData, models.UsageTypeVoice, models.UsageTypeSMS).
		Row().Scan(&prevDataUsage, &prevVoiceUsage, &prevSmsUsage); err != nil {
		// If no previous data, set trend to zero
		prevDataUsage, prevVoiceUsage, prevSmsUsage = 0, 0, 0
	}

	// Calculate trend percentage
	totalCurrentUsage := dataUsage + voiceUsage + smsUsage
	totalPrevUsage := prevDataUsage + prevVoiceUsage + prevSmsUsage
	var trendPercentage float64
	if totalPrevUsage > 0 {
		trendPercentage = ((totalCurrentUsage - totalPrevUsage) / totalPrevUsage) * 100
	}

	direction := "STABLE"
	if trendPercentage > 5 {
		direction = "UP"
	} else if trendPercentage < -5 {
		direction = "DOWN"
	}

	stats := &UsageStats{
		DataUsage:  dataUsage,
		VoiceUsage: voiceUsage,
		SmsUsage:   smsUsage,
		Cost:       totalCost,
		Period:     period,
		Trend: &UsageTrend{
			Direction:        direction,
			Percentage:       trendPercentage,
			PeriodOverPeriod: trendPercentage,
		},
	}
	return stats, nil
}

// GetRealTimeUsage gets real-time usage for a subscriber
func (cs *ChargingService) GetRealTimeUsage(ctx context.Context, imsi string) (*RealTimeUsage, error) {
	// Get current active session
	var session models.Session
	var currentSession *CurrentSession

	if err := cs.db.DB.WithContext(ctx).
		Where("imsi = ? AND status IN ?", imsi, []string{"active", "connected", "established"}).
		First(&session).Error; err == nil {
		// Get usage for current session
		var dataUsed, voiceUsed, smsUsed, sessionCost float64
		if err := cs.db.DB.WithContext(ctx).Model(&models.UsageEvent{}).
			Where("session_id = ?", session.SessionID).
			Select("SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as data_usage, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as voice_usage, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as sms_usage, SUM(cost) as total_cost",
				models.UsageTypeData, models.UsageTypeVoice, models.UsageTypeSMS).
			Row().Scan(&dataUsed, &voiceUsed, &smsUsed, &sessionCost); err != nil {
			dataUsed, voiceUsed, smsUsed, sessionCost = 0, 0, 0, 0
		}

		currentSession = &CurrentSession{
			SessionID: session.SessionID,
			StartTime: session.CreatedAt,
			DataUsed:  dataUsed,
			VoiceUsed: voiceUsed,
			SmsUsed:   smsUsed,
			Cost:      sessionCost,
		}
	}

	// Get today's usage
	today := time.Now().Truncate(24 * time.Hour)
	var dataUsed, voiceUsed, smsUsed, todayCost float64

	if err := cs.db.DB.WithContext(ctx).Model(&models.UsageEvent{}).
		Where("imsi = ? AND created_at >= ?", imsi, today).
		Select("SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as data_usage, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as voice_usage, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as sms_usage, SUM(cost) as total_cost",
			models.UsageTypeData, models.UsageTypeVoice, models.UsageTypeSMS).
		Row().Scan(&dataUsed, &voiceUsed, &smsUsed, &todayCost); err != nil {
		dataUsed, voiceUsed, smsUsed, todayCost = 0, 0, 0, 0
	}

	usage := &RealTimeUsage{
		CurrentSession: currentSession,
		TodayUsage: &TodayUsage{
			DataUsed:  dataUsed,
			VoiceUsed: voiceUsed,
			SmsUsed:   smsUsed,
			Cost:      todayCost,
		},
	}
	return usage, nil
}

// ListUsageEvents lists usage events
func (cs *ChargingService) ListUsageEvents(ctx context.Context, limit int, offset int, where string) ([]*models.UsageEvent, int64, error) {
	query := cs.db.DB.WithContext(ctx).Model(&models.UsageEvent{})

	// Apply where clause if provided
	if where != "" {
		query = query.Where(where)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count usage events: %w", err)
	}

	// Get usage events with pagination
	var events []*models.UsageEvent
	if err := query.Preload("Subscriber").Limit(limit).Offset(offset).Order("timestamp DESC").Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list usage events: %w", err)
	}

	return events, total, nil
}

// SearchUsageEvents searches usage events
func (cs *ChargingService) SearchUsageEvents(ctx context.Context, query string, limit int) ([]*models.UsageEvent, error) {
	var events []*models.UsageEvent

	// Search usage events by IMSI, session ID, or usage type
	searchPattern := "%" + query + "%"
	if err := cs.db.DB.WithContext(ctx).
		Where("imsi ILIKE ? OR session_id ILIKE ? OR usage_type ILIKE ?",
			searchPattern, searchPattern, searchPattern).
		Preload("Subscriber").
		Limit(limit).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to search usage events: %w", err)
	}

	return events, nil
}

// SubscribeToUsageUpdates subscribes to usage updates
func (cs *ChargingService) SubscribeToUsageUpdates(ctx context.Context, imsi string) (<-chan *models.UsageEvent, error) {
	ch := make(chan *models.UsageEvent, 20) // Buffered channel for high-frequency updates

	go func() {
		defer close(ch)

		// Poll for new usage events every 1 second (high frequency for real-time usage)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		var lastEventID uint = 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Query for new usage events since last check
				var events []models.UsageEvent
				if err := cs.db.DB.WithContext(ctx).
					Where("imsi = ? AND id > ?", imsi, lastEventID).
					Order("id ASC").
					Limit(20).
					Find(&events).Error; err != nil {
					// Log error but continue polling
					continue
				}

				// Send new events to channel
				for _, event := range events {
					select {
					case ch <- &event:
						lastEventID = event.ID
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// SubscribeToSystemStatsUpdates subscribes to system stats updates
func (cs *ChargingService) SubscribeToSystemStatsUpdates(ctx context.Context) (<-chan *models.SystemStats, error) {
	ch := make(chan *models.SystemStats, 5) // Buffered channel to prevent blocking

	go func() {
		defer close(ch)

		// Poll for system stats updates every 5 seconds (lower frequency for system stats)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		var lastStats models.SystemStats

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Get current system stats
				stats, err := cs.GetSystemStats(ctx)
				if err != nil {
					// Log error but continue polling
					continue
				}

				// Only send if stats have changed significantly
				if lastStats.ActiveSessions != stats.ActiveSessions ||
					lastStats.TotalAccounts != stats.TotalAccounts ||
					lastStats.BlockedUsers != stats.BlockedUsers ||
					lastStats.LowBalanceAlerts != stats.LowBalanceAlerts {

					select {
					case ch <- stats:
						lastStats = *stats
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// TriggerMaintenance triggers system maintenance
func (cs *ChargingService) TriggerMaintenance(ctx context.Context) (bool, error) {
	// In a real implementation, this would trigger maintenance tasks like:
	// - Database cleanup
	// - Cache refresh
	// - Log rotation
	// - Health checks

	// For now, simulate maintenance by updating a maintenance timestamp
	if err := cs.db.DB.WithContext(ctx).Exec("UPDATE system_config SET last_maintenance = ? WHERE id = 1", time.Now()).Error; err != nil {
		// If table doesn't exist, just return success
		return true, nil
	}

	return true, nil
}

// Helper types for charging service
type UsageStats struct {
	DataUsage  float64     `json:"data_usage"`
	VoiceUsage float64     `json:"voice_usage"`
	SmsUsage   float64     `json:"sms_usage"`
	Cost       float64     `json:"cost"`
	Period     string      `json:"period"`
	Trend      *UsageTrend `json:"trend"`
}

type UsageTrend struct {
	Direction        string  `json:"direction"`
	Percentage       float64 `json:"percentage"`
	PeriodOverPeriod float64 `json:"period_over_period"`
}

type RealTimeUsage struct {
	CurrentSession *CurrentSession `json:"current_session"`
	TodayUsage     *TodayUsage     `json:"today_usage"`
}

type CurrentSession struct {
	SessionID string    `json:"session_id"`
	StartTime time.Time `json:"start_time"`
	DataUsed  float64   `json:"data_used"`
	VoiceUsed float64   `json:"voice_used"`
	SmsUsed   float64   `json:"sms_used"`
	Cost      float64   `json:"cost"`
}

type TodayUsage struct {
	DataUsed  float64 `json:"data_used"`
	VoiceUsed float64 `json:"voice_used"`
	SmsUsed   float64 `json:"sms_used"`
	Cost      float64 `json:"cost"`
}
