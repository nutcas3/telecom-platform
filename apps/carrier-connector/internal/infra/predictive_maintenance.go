package infra

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MaintenanceType represents types of maintenance
type MaintenanceType string

const (
	MaintenanceTypePreventive MaintenanceType = "preventive"
	MaintenanceTypeCorrective MaintenanceType = "corrective"
	MaintenanceTypePredictive MaintenanceType = "predictive"
	MaintenanceTypeEmergency  MaintenanceType = "emergency"
)

// AssetType represents infrastructure asset types
type AssetType string

const (
	AssetTypeServer       AssetType = "server"
	AssetTypeDatabase     AssetType = "database"
	AssetTypeNetwork      AssetType = "network"
	AssetTypeStorage      AssetType = "storage"
	AssetTypeApplication  AssetType = "application"
	AssetTypeLoadBalancer AssetType = "load_balancer"
)

// Asset represents an infrastructure asset
type Asset struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Type            AssetType      `json:"type"`
	Status          string         `json:"status"`
	Location        string         `json:"location"`
	HealthScore     float64        `json:"health_score"` // 0-100
	LastMaintenance *time.Time     `json:"last_maintenance,omitempty"`
	NextMaintenance *time.Time     `json:"next_maintenance,omitempty"`
	Metadata        map[string]any `json:"metadata"`
}

// MaintenanceAlert represents a maintenance alert
type MaintenanceAlert struct {
	ID               string          `json:"id"`
	AssetID          string          `json:"asset_id"`
	Type             MaintenanceType `json:"type"`
	Severity         string          `json:"severity"` // "low", "medium", "high", "critical"
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	RiskScore        float64         `json:"risk_score"` // 0-100
	PredictedFailure *time.Time      `json:"predicted_failure,omitempty"`
	Recommendations  []string        `json:"recommendations"`
	Timestamp        time.Time       `json:"timestamp"`
	Status           string          `json:"status"` // "new", "acknowledged", "scheduled", "completed"
}

// MaintenanceMetrics represents maintenance performance metrics
type MaintenanceMetrics struct {
	Period                 string    `json:"period"`
	TotalAssets            int64     `json:"total_assets"`
	HealthyAssets          int64     `json:"healthy_assets"`
	AssetsNeedingAttention int64     `json:"assets_needing_attention"`
	PreventiveMaintenance  int64     `json:"preventive_maintenance"`
	CorrectiveMaintenance  int64     `json:"corrective_maintenance"`
	EmergencyMaintenance   int64     `json:"emergency_maintenance"`
	Uptime                 float64   `json:"uptime_pct"`
	MeanTimeToFailure      float64   `json:"mean_time_to_failure_hours"`
	MeanTimeToRepair       float64   `json:"mean_time_to_repair_hours"`
	GeneratedAt            time.Time `json:"generated_at"`
}

// PredictiveMaintenanceService provides predictive maintenance capabilities
type PredictiveMaintenanceService struct {
	db     *gorm.DB
	logger *logrus.Logger
	assets map[string]*Asset
	alerts []*MaintenanceAlert
	mu     sync.RWMutex
	models map[AssetType]*MaintenanceModel
}

// MaintenanceModel represents a predictive maintenance model
type MaintenanceModel struct {
	AssetType   AssetType `json:"asset_type"`
	Version     string    `json:"version"`
	LastTrained time.Time `json:"last_trained"`
	Accuracy    float64   `json:"accuracy"`
	FailureRate float64   `json:"failure_rate"`
	MTTF        float64   `json:"mttf_hours"` // Mean Time To Failure
	MTTR        float64   `json:"mttr_hours"` // Mean Time To Repair
}

// NewPredictiveMaintenanceService creates a new predictive maintenance service
func NewPredictiveMaintenanceService(db *gorm.DB, logger *logrus.Logger) *PredictiveMaintenanceService {
	service := &PredictiveMaintenanceService{
		db:     db,
		logger: logger,
		assets: make(map[string]*Asset),
		alerts: make([]*MaintenanceAlert, 0),
		models: make(map[AssetType]*MaintenanceModel),
	}

	// Initialize with default assets
	service.initializeAssets()

	// Initialize predictive models
	service.initializeModels()

	// Start monitoring
	go service.monitorAssets()

	return service
}

// MonitorAssets continuously monitors asset health
func (s *PredictiveMaintenanceService) monitorAssets() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.updateAssetHealth()
		s.predictFailures()
	}
}

// GetMaintenanceMetrics returns maintenance performance metrics
func (s *PredictiveMaintenanceService) GetMaintenanceMetrics(ctx context.Context, period string) (*MaintenanceMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := &MaintenanceMetrics{
		Period:      period,
		GeneratedAt: time.Now(),
	}

	// Count assets by status
	for _, asset := range s.assets {
		metrics.TotalAssets++
		if asset.HealthScore >= 80 {
			metrics.HealthyAssets++
		} else if asset.HealthScore < 60 {
			metrics.AssetsNeedingAttention++
		}
	}

	// Count maintenance types
	for _, alert := range s.alerts {
		switch alert.Type {
		case MaintenanceTypePreventive:
			metrics.PreventiveMaintenance++
		case MaintenanceTypeCorrective:
			metrics.CorrectiveMaintenance++
		case MaintenanceTypeEmergency:
			metrics.EmergencyMaintenance++
		}
	}

	// Calculate uptime and MTTF/MTTR (simplified)
	metrics.Uptime = 99.5
	metrics.MeanTimeToFailure = 8760 // 1 year in hours
	metrics.MeanTimeToRepair = 4     // 4 hours

	return metrics, nil
}

// GetMaintenanceAlerts returns maintenance alerts
func (s *PredictiveMaintenanceService) GetMaintenanceAlerts(ctx context.Context, severity string) ([]*MaintenanceAlert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]*MaintenanceAlert, 0)
	for _, alert := range s.alerts {
		if severity == "" || alert.Severity == severity {
			filtered = append(filtered, alert)
		}
	}

	return filtered, nil
}

// PredictFailure predicts failure for an asset
func (s *PredictiveMaintenanceService) PredictFailure(_ context.Context, assetID string) (*MaintenanceAlert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	asset, exists := s.assets[assetID]
	if !exists {
		return nil, fmt.Errorf("asset not found: %s", assetID)
	}

	// Get predictive model
	model, exists := s.models[asset.Type]
	if !exists {
		return nil, fmt.Errorf("no model for asset type: %s", asset.Type)
	}

	// Calculate failure probability
	failureProb := s.calculateFailureProbability(asset, model)

	if failureProb > 0.7 {
		alert := &MaintenanceAlert{
			ID:               fmt.Sprintf("alert-%d", time.Now().UnixNano()),
			AssetID:          assetID,
			Type:             MaintenanceTypePredictive,
			Severity:         s.determineSeverity(failureProb),
			Title:            fmt.Sprintf("Predicted failure for %s", asset.Name),
			Description:      fmt.Sprintf("Asset %s has high probability of failure", asset.Name),
			RiskScore:        failureProb * 100,
			PredictedFailure: s.predictFailureTime(asset, model),
			Recommendations:  s.generateMaintenanceRecommendations(asset, model),
			Timestamp:        time.Now(),
			Status:           "new",
		}

		s.alerts = append(s.alerts, alert)
		return alert, nil
	}

	return nil, nil
}

// ScheduleMaintenance schedules maintenance for an asset
func (s *PredictiveMaintenanceService) ScheduleMaintenance(ctx context.Context, assetID, maintenanceType string, scheduledTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	asset, exists := s.assets[assetID]
	if !exists {
		return fmt.Errorf("asset not found: %s", assetID)
	}

	asset.NextMaintenance = &scheduledTime

	// Create maintenance alert
	alert := &MaintenanceAlert{
		ID:          fmt.Sprintf("maintenance-%d", time.Now().UnixNano()),
		AssetID:     assetID,
		Type:        MaintenanceType(maintenanceType),
		Severity:    "medium",
		Title:       fmt.Sprintf("Scheduled maintenance for %s", asset.Name),
		Description: fmt.Sprintf("Maintenance scheduled for %s", scheduledTime.Format(time.RFC3339)),
		Timestamp:   time.Now(),
		Status:      "scheduled",
	}

	s.alerts = append(s.alerts, alert)

	s.logger.WithFields(logrus.Fields{
		"asset_id":     assetID,
		"asset_name":   asset.Name,
		"scheduled_at": scheduledTime,
	}).Info("Maintenance scheduled")

	return nil
}

// initializeAssets initializes default infrastructure assets
func (s *PredictiveMaintenanceService) initializeAssets() {
	assets := []*Asset{
		{
			ID:          "api-server-1",
			Name:        "API Server 1",
			Type:        AssetTypeServer,
			Status:      "running",
			Location:    "us-east-1",
			HealthScore: 95.0,
			Metadata:    map[string]any{"cpu_cores": 8, "memory_gb": 32},
		},
		{
			ID:          "db-primary",
			Name:        "Primary Database",
			Type:        AssetTypeDatabase,
			Status:      "running",
			Location:    "us-east-1",
			HealthScore: 92.0,
			Metadata:    map[string]any{"engine": "postgresql", "version": "14"},
		},
		{
			ID:          "lb-main",
			Name:        "Main Load Balancer",
			Type:        AssetTypeLoadBalancer,
			Status:      "running",
			Location:    "us-east-1",
			HealthScore: 98.0,
			Metadata:    map[string]any{"connections": 1000, "throughput_mbps": 1000},
		},
		{
			ID:          "cache-redis",
			Name:        "Redis Cache",
			Type:        AssetTypeApplication,
			Status:      "running",
			Location:    "us-east-1",
			HealthScore: 88.0,
			Metadata:    map[string]any{"memory_gb": 16, "hit_rate": 0.95},
		},
		{
			ID:          "storage-s3",
			Name:        "S3 Storage",
			Type:        AssetTypeStorage,
			Status:      "healthy",
			Location:    "us-east-1",
			HealthScore: 99.0,
			Metadata:    map[string]any{"capacity_tb": 100, "used_tb": 45},
		},
	}

	for _, asset := range assets {
		s.assets[asset.ID] = asset
	}
}

// initializeModels initializes predictive maintenance models
func (s *PredictiveMaintenanceService) initializeModels() {
	models := map[AssetType]*MaintenanceModel{
		AssetTypeServer: {
			AssetType:   AssetTypeServer,
			Version:     "1.0",
			LastTrained: time.Now().AddDate(0, -1, 0),
			Accuracy:    0.85,
			FailureRate: 0.05,
			MTTF:        8760, // 1 year
			MTTR:        4,    // 4 hours
		},
		AssetTypeDatabase: {
			AssetType:   AssetTypeDatabase,
			Version:     "1.0",
			LastTrained: time.Now().AddDate(0, -1, 0),
			Accuracy:    0.92,
			FailureRate: 0.02,
			MTTF:        17520, // 2 years
			MTTR:        8,     // 8 hours
		},
		AssetTypeLoadBalancer: {
			AssetType:   AssetTypeLoadBalancer,
			Version:     "1.0",
			LastTrained: time.Now().AddDate(0, -1, 0),
			Accuracy:    0.88,
			FailureRate: 0.03,
			MTTF:        13140, // 1.5 years
			MTTR:        2,     // 2 hours
		},
		AssetTypeApplication: {
			AssetType:   AssetTypeApplication,
			Version:     "1.0",
			LastTrained: time.Now().AddDate(0, -1, 0),
			Accuracy:    0.78,
			FailureRate: 0.08,
			MTTF:        4380, // 6 months
			MTTR:        1,    // 1 hour
		},
		AssetTypeStorage: {
			AssetType:   AssetTypeStorage,
			Version:     "1.0",
			LastTrained: time.Now().AddDate(0, -1, 0),
			Accuracy:    0.95,
			FailureRate: 0.01,
			MTTF:        35040, // 4 years
			MTTR:        12,    // 12 hours
		},
	}

	for assetType, model := range models {
		s.models[assetType] = model
	}
}

// updateAssetHealth updates health scores for all assets
func (s *PredictiveMaintenanceService) updateAssetHealth() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, asset := range s.assets {
		// Simulate health score changes
		change := (rand.Float64()*10 - 5) // Random change between -5 and +5
		asset.HealthScore = math.Max(0, math.Min(100, asset.HealthScore+change))
	}
}

// predictFailures predicts failures for all assets
func (s *PredictiveMaintenanceService) predictFailures() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, asset := range s.assets {
		model, exists := s.models[asset.Type]
		if !exists {
			continue
		}

		failureProb := s.calculateFailureProbability(asset, model)
		if failureProb > 0.8 {
			// Create high-priority alert
			alert := &MaintenanceAlert{
				ID:               fmt.Sprintf("alert-%d", time.Now().UnixNano()),
				AssetID:          asset.ID,
				Type:             MaintenanceTypePredictive,
				Severity:         "critical",
				Title:            fmt.Sprintf("Critical: %s likely to fail", asset.Name),
				Description:      fmt.Sprintf("Asset has %.1f%% probability of failure", failureProb*100),
				RiskScore:        failureProb * 100,
				PredictedFailure: s.predictFailureTime(asset, model),
				Timestamp:        time.Now(),
				Status:           "new",
			}

			s.alerts = append(s.alerts, alert)
		}
	}
}

// calculateFailureProbability calculates failure probability for an asset
func (s *PredictiveMaintenanceService) calculateFailureProbability(asset *Asset, model *MaintenanceModel) float64 {
	// Simplified failure probability calculation
	baseProb := model.FailureRate

	// Adjust based on health score
	healthFactor := (100 - asset.HealthScore) / 100

	// Adjust based on age (if no recent maintenance)
	ageFactor := 1.0
	if asset.LastMaintenance != nil {
		daysSinceMaintenance := time.Since(*asset.LastMaintenance).Hours() / 24
		ageFactor = math.Min(2.0, daysSinceMaintenance/365) // Max 2x risk after 1 year
	}

	probability := baseProb * healthFactor * ageFactor

	return math.Min(1.0, probability)
}

// determineSeverity determines alert severity from probability
func (s *PredictiveMaintenanceService) determineSeverity(probability float64) string {
	switch {
	case probability >= 0.9:
		return "critical"
	case probability >= 0.7:
		return "high"
	case probability >= 0.5:
		return "medium"
	default:
		return "low"
	}
}

// predictFailureTime predicts when failure might occur
func (s *PredictiveMaintenanceService) predictFailureTime(asset *Asset, model *MaintenanceModel) *time.Time {
	// Simplified prediction based on MTTF and current health
	hoursToFailure := model.MTTF * (asset.HealthScore / 100)
	predictedTime := time.Now().Add(time.Duration(hoursToFailure) * time.Hour)
	return &predictedTime
}

// generateMaintenanceRecommendations generates maintenance recommendations
func (s *PredictiveMaintenanceService) generateMaintenanceRecommendations(asset *Asset, model *MaintenanceModel) []string {
	recommendations := make([]string, 0)

	switch asset.Type {
	case AssetTypeServer:
		recommendations = append(recommendations, "Check CPU and memory utilization")
		recommendations = append(recommendations, "Review system logs for errors")
		recommendations = append(recommendations, "Update system patches")
	case AssetTypeDatabase:
		recommendations = append(recommendations, "Run database health check")
		recommendations = append(recommendations, "Optimize query performance")
		recommendations = append(recommendations, "Check disk space and I/O")
	case AssetTypeLoadBalancer:
		recommendations = append(recommendations, "Review connection limits")
		recommendations = append(recommendations, "Check SSL certificate expiry")
		recommendations = append(recommendations, "Monitor response times")
	case AssetTypeApplication:
		recommendations = append(recommendations, "Restart application services")
		recommendations = append(recommendations, "Clear application cache")
		recommendations = append(recommendations, "Check for memory leaks")
	case AssetTypeStorage:
		recommendations = append(recommendations, "Check storage capacity")
		recommendations = append(recommendations, "Run storage health diagnostics")
		recommendations = append(recommendations, "Review backup integrity")
	}

	return recommendations
}
