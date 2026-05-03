package integration

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/handlers"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/services"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
	"github.com/sirupsen/logrus"
)

type SelectionIntegration struct {
	manager          *smdp.SMDPManager
	selectionService *services.SelectionService
	selectionHandler *handlers.SelectionHandler
	smdpService      *services.SMDPService
	server           *http.Server
}

func NewSelectionIntegration(repo *repository.PostgresProfileStore) *SelectionIntegration {
	// Create SMDP manager with default configuration
	config := &smdp.ManagerConfig{
		HealthCheckInterval: 30 * time.Second,
		EnableFailover:      true,
		EnableLoadBalancing: true,
		MaxRetries:          3,
		RetryDelay:          1 * time.Second,
	}

	manager := smdp.NewSMDPManager(repo, config)

	// Create selection service
	logger := logrus.New()
	selectionService := services.NewSelectionService(manager, logger)

	// Create SMDP service
	smdpService := services.NewSMDPService(repo)

	return &SelectionIntegration{
		manager:          manager,
		selectionService: selectionService,
		selectionHandler: handlers.NewSelectionHandler(manager),
		smdpService:      smdpService,
	}
}

// StartServer starts the HTTP server with all endpoints
func (si *SelectionIntegration) StartServer(port string) error {
	// Create HTTP multiplexer
	mux := http.NewServeMux()

	// Register selection routes
	si.selectionHandler.RegisterRoutes(mux)

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "selection-integration"}`))
	})

	// Create and configure server
	si.server = &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Starting selection integration server on port %s", port)
	return si.server.ListenAndServe()
}

// StartHealthChecking starts the background health checking process
func (si *SelectionIntegration) StartHealthChecking(ctx context.Context) {
	si.manager.StartHealthChecking(ctx)
}

// GetManager returns the SMDP manager for testing
func (si *SelectionIntegration) GetManager() *smdp.SMDPManager {
	return si.manager
}

// GetSelectionService returns the selection service for testing
func (si *SelectionIntegration) GetSelectionService() *services.SelectionService {
	return si.selectionService
}

// Shutdown gracefully shuts down the server
func (si *SelectionIntegration) Shutdown(ctx context.Context) error {
	if si.server != nil {
		return si.server.Shutdown(ctx)
	}
	return nil
}

func (si *SelectionIntegration) RunDemo(ctx context.Context) error {
	log.Println("Starting carrier selection demonstration...")

	if err := si.SetupCarriers(); err != nil {
		return err
	}

	si.StartHealthChecking(ctx)

	time.Sleep(2 * time.Second)

	log.Println("Demonstrating intelligent carrier selection...")

	scenarios := []struct {
		name     string
		criteria *smdp.SelectionCriteria
	}{
		{
			name: "High Priority US Request",
			criteria: &smdp.SelectionCriteria{
				Region:            "US",
				ProfileType:       "operational",
				Urgency:           "high",
				CostSensitivity:   0.2,
				PerformanceWeight: 0.6,
				ReliabilityWeight: 0.6,
			},
		},
		{
			name: "Cost-Optimized European Request",
			criteria: &smdp.SelectionCriteria{
				Region:            "DE",
				ProfileType:       "operational",
				Urgency:           "low",
				CostSensitivity:   0.8,
				PerformanceWeight: 0.2,
				ReliabilityWeight: 0.3,
			},
		},
		{
			name: "Balanced Global Request",
			criteria: &smdp.SelectionCriteria{
				Region:            "",
				ProfileType:       "operational",
				Urgency:           "medium",
				CostSensitivity:   0.5,
				PerformanceWeight: 0.4,
				ReliabilityWeight: 0.4,
			},
		},
	}

	for _, scenario := range scenarios {
		log.Printf("Testing scenario: %s", scenario.name)

		score, err := si.manager.SelectOptimalCarrier(ctx, scenario.criteria)
		if err != nil {
			log.Printf("Error in scenario %s: %v", scenario.name, err)
			continue
		}

		log.Printf("Selected carrier: %s (%s)", score.Carrier.Name, score.CarrierID)
		log.Printf("Total score: %.2f", score.TotalScore)
		log.Printf("Performance: %.2f, Reliability: %.2f, Cost: %.2f, Region: %.2f, Capability: %.2f",
			score.PerformanceScore, score.ReliabilityScore, score.CostScore, score.RegionScore, score.CapabilityScore)
		log.Printf("Reason: %s", score.Reason)
		log.Println("---")
	}

	log.Println("Generating carrier analytics...")
	analytics, err := si.selectionService.GetCarrierAnalytics(ctx)
	if err != nil {
		log.Printf("Error generating analytics: %v", err)
	} else {
		log.Printf("Analytics generated for %d carriers", analytics.TotalCarriers)
		log.Printf("Overall health: %.1f%%", analytics.Summary.OverallHealth)
		log.Printf("Overall success rate: %.1f%%", analytics.Summary.OverallSuccessRate)
	}

	log.Println("Generating optimization recommendations...")
	optimization, err := si.selectionService.OptimizeCarrierSelection(ctx)
	if err != nil {
		log.Printf("Error generating optimization: %v", err)
	} else {
		log.Printf("System health: %.1f%%", optimization.OverallHealth)
		log.Printf("Recommendations: %d", len(optimization.Recommendations))
		log.Printf("Priority actions: %d", len(optimization.PriorityActions))
	}

	log.Println("Carrier selection demonstration completed successfully!")
	return nil
}
