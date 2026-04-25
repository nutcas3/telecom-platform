package main

import (
	"log"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/handlers"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/metrics"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/monitoring"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/rbac"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// serverDeps bundles the services and handlers needed to register routes.
type serverDeps struct {
	authSvc   *services.AuthService
	casbinSvc *rbac.CasbinService

	authH        *handlers.AuthHandler
	servicesH    *handlers.ServicesHandler
	monitoringH  *handlers.MonitoringHandler
	deploymentsH *handlers.DeploymentsHandler
	pluginsH     *handlers.PluginsHandler
	automationH  *handlers.AutomationHandler
	configH      *handlers.ConfigHandler
	chaosH       *handlers.ChaosHandler
	billingH     *handlers.BillingHandler
	chargingH    *handlers.ChargingHandler
}

// buildDeps constructs all downstream services and handlers.
func buildDeps(db *database.Database, cfg *config.Config) *serverDeps {
	authSvc := services.NewAuthService(db.DB, cfg.Auth.JWTSecret)

	casbinSvc, err := rbac.NewCasbinService(db.DB)
	if err != nil {
		log.Printf("Failed to initialize Casbin service: %v", err)
		casbinSvc = nil
	}

	chaosSvc := services.NewChaosService(db)
	invoiceSvc := services.NewInvoiceService(db)
	pluginSvc := services.NewPluginService(db)
	automationSvc := services.NewAutomationService(db)
	configStoreSvc := services.NewConfigStoreService(db)
	deploymentSvc := services.NewDeploymentService(db)
	promSvc := monitoring.NewPrometheusService()

	// Initialize charging engine client
	chargingEngineClient := services.NewChargingEngineClient(&cfg.ChargingEngine)

	k8sSvc, k8sErr := services.NewKubernetesService()
	if k8sErr != nil {
		log.Printf("Kubernetes integration disabled: %v", k8sErr)
	}

	return &serverDeps{
		authSvc:      authSvc,
		casbinSvc:    casbinSvc,
		authH:        handlers.NewAuthHandler(authSvc),
		servicesH:    handlers.NewServicesHandler(k8sSvc),
		monitoringH:  handlers.NewMonitoringHandler(promSvc),
		deploymentsH: handlers.NewDeploymentsHandler(deploymentSvc),
		pluginsH:     handlers.NewPluginsHandler(pluginSvc),
		automationH:  handlers.NewAutomationHandler(automationSvc),
		configH:      handlers.NewConfigHandler(configStoreSvc),
		chaosH:       handlers.NewChaosHandler(chaosSvc),
		billingH:     handlers.NewBillingHandler(invoiceSvc, db.DB),
		chargingH:    handlers.NewChargingHandler(chargingEngineClient),
	}
}

// buildMetricsCollector constructs the Prometheus collector, optionally wiring
// the InfluxDB time-series storage when configured.
func buildMetricsCollector(cfg *config.Config) (*metrics.MetricsCollector, *metrics.MetricsCollectorWithTimeSeries, *metrics.TimeSeriesStorage) {
	prometheusCollector := metrics.NewMetricsCollector()
	prometheusCollector.Register()

	var timeSeriesStorage *metrics.TimeSeriesStorage
	if cfg.InfluxDB.URL != "" {
		ts, err := metrics.NewTimeSeriesStorage(
			cfg.InfluxDB.URL,
			cfg.InfluxDB.Token,
			cfg.InfluxDB.Org,
			cfg.InfluxDB.Bucket,
		)
		if err != nil {
			log.Printf("Failed to initialize time-series storage: %v", err)
		} else {
			timeSeriesStorage = ts
		}
	}

	var withTS *metrics.MetricsCollectorWithTimeSeries
	if timeSeriesStorage != nil {
		withTS = metrics.NewMetricsCollectorWithTimeSeries(prometheusCollector, timeSeriesStorage)
	} else {
		withTS = &metrics.MetricsCollectorWithTimeSeries{MetricsCollector: prometheusCollector}
	}

	return prometheusCollector, withTS, timeSeriesStorage
}
