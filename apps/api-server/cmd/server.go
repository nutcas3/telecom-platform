package main

import (
	"context"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/graphql"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/handlers"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/metrics"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/middleware"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/monitoring"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/rbac"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/websocket"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize metrics collector
	prometheusCollector := metrics.NewMetricsCollector()
	prometheusCollector.Register()

	// Initialize time-series storage (optional)
	var timeSeriesStorage *metrics.TimeSeriesStorage
	if cfg.InfluxDB.URL != "" {
		timeSeriesStorage, err = metrics.NewTimeSeriesStorage(
			cfg.InfluxDB.URL,
			cfg.InfluxDB.Token,
			cfg.InfluxDB.Org,
			cfg.InfluxDB.Bucket,
		)
		if err != nil {
			log.Printf("Failed to initialize time-series storage: %v", err)
		}
		defer func() {
			if timeSeriesStorage != nil {
				timeSeriesStorage.Close()
			}
		}()
	}

	// Create metrics collector with time-series
	var metricsCollector *metrics.MetricsCollectorWithTimeSeries
	if timeSeriesStorage != nil {
		metricsCollector = metrics.NewMetricsCollectorWithTimeSeries(prometheusCollector, timeSeriesStorage)
	} else {
		metricsCollector = &metrics.MetricsCollectorWithTimeSeries{
			MetricsCollector: prometheusCollector,
		}
	}

	// Initialize GraphQL resolver
	resolver := graphql.NewResolver(db, cfg)

	// Setup HTTP server with GraphQL handler
	router := gin.Default()
	graphql.SetupGraphQLHandler(router, resolver)

	// Add metrics middleware
	router.Use(func(c *gin.Context) {
		// Create a wrapper handler that calls the next handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})

		// Apply prometheus middleware
		middlewareHandler := prometheusCollector.HTTPMiddleware(next)
		middlewareHandler.ServeHTTP(c.Writer, c.Request)
	})

	// Add metrics endpoint
	router.GET("/metrics", gin.WrapH(prometheusCollector.Handler()))

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api-server",
		})
	})

	// Add Swagger documentation endpoints
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Add WebSocket endpoint
	router.GET("/ws", websocket.HandleWebSocket(websocket.GetHub()))

	// Add health check endpoints
	router.GET("/health", monitoring.HealthHandler())
	router.GET("/ready", monitoring.ReadyHandler())
	router.GET("/live", monitoring.LiveHandler())

	// Apply global rate limiting
	router.Use(middleware.RateLimitByEndpoint())

	// Apply performance optimization middleware
	router.Use(middleware.PerformanceMiddlewareHandler())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestSizeMiddleware(10 << 20)) // 10MB limit
	router.Use(middleware.TimeoutMiddleware(30 * time.Second))

	// Add performance metrics endpoint
	router.GET("/metrics/performance", middleware.PerformanceHandler())

	// Initialize authentication service
	authSvc := services.NewAuthService(db.DB, cfg.Auth.JWTSecret)

	// Initialize WebSocket hub
	websocket.InitializeWebSocket()

	// Initialize health monitor
	monitoring.InitializeHealthMonitor("1.0.0", "production")

	// Register health checks
	healthMonitor := monitoring.GetHealthMonitor()
	healthMonitor.RegisterCheck("database", monitoring.NewDatabaseHealthChecker(db.DB))

	// Register system health check
	healthMonitor.RegisterCheck("system", monitoring.NewSystemHealthChecker())

	// Initialize alert manager
	monitoring.InitializeAlertManager(healthMonitor)

	// Initialize performance middleware
	middleware.InitializePerformanceMiddleware(100 * time.Millisecond) // 100ms threshold

	// Initialize Casbin RBAC service
	casbinSvc, err := rbac.NewCasbinService(db.DB)
	if err != nil {
		log.Printf("Failed to initialize Casbin service: %v", err)
		// Continue without Casbin - fallback to basic role checking
		casbinSvc = nil
	}

	// Initialize real, DB/K8s/Prometheus-backed services.
	chaosSvc := services.NewChaosService(db)
	invoiceSvc := services.NewInvoiceService(db)
	pluginSvc := services.NewPluginService(db)
	automationSvc := services.NewAutomationService(db)
	configStoreSvc := services.NewConfigStoreService(db)
	deploymentSvc := services.NewDeploymentService(db)
	promSvc := monitoring.NewPrometheusService()

	// Kubernetes service is optional: if not configured we surface a clear
	// 503 from service management endpoints rather than returning fake data.
	k8sSvc, k8sErr := services.NewKubernetesService()
	if k8sErr != nil {
		log.Printf("Kubernetes integration disabled: %v", k8sErr)
	}

	// Build handlers wired to real services.
	authH := handlers.NewAuthHandler(authSvc)
	servicesH := handlers.NewServicesHandler(k8sSvc)
	monitoringH := handlers.NewMonitoringHandler(promSvc)
	deploymentsH := handlers.NewDeploymentsHandler(deploymentSvc)
	pluginsH := handlers.NewPluginsHandler(pluginSvc)
	automationH := handlers.NewAutomationHandler(automationSvc)
	configH := handlers.NewConfigHandler(configStoreSvc)
	chaosH := handlers.NewChaosHandler(chaosSvc)
	billingH := handlers.NewBillingHandler(invoiceSvc, db.DB)

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Authentication routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authH.Login)
			auth.POST("/register", authH.Register)
			auth.POST("/refresh", authH.RefreshToken)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(authSvc))
		{
			// User management
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", authH.Logout)
				authProtected.GET("/profile", authH.GetProfile)
				authProtected.POST("/change-password", authH.ChangePassword)
			}

			// Admin-only user management
			users := protected.Group("/users")
			if casbinSvc != nil {
				users.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/users", "GET"))
			} else {
				users.Use(middleware.RequireRole("admin"))
			}
			{
				users.GET("", authH.GetUsers)
				users.POST("", authH.CreateUser)
				users.PUT("/:id", authH.UpdateUser)
				users.DELETE("/:id", authH.DeleteUser)
			}
		}

		// Protected API endpoints (authentication required)
		apiProtected := v1.Group("/")
		apiProtected.Use(middleware.AuthMiddleware(authSvc))
		{
			// Services Management
			svcs := apiProtected.Group("/services")
			{
				svcs.GET("", servicesH.List)
				svcs.GET("/:id", servicesH.Get)
				svcs.GET("/:id/logs", servicesH.Logs)
				svcs.GET("/:id/health", servicesH.Health)
				svcs.GET("/:id/pods", servicesH.PodStatus)
				svcs.GET("/:id/events", servicesH.Events)

				// Write operations require higher permissions
				svcsWrite := svcs.Group("/")
				if casbinSvc != nil {
					svcsWrite.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/services", "POST"))
				} else {
					svcsWrite.Use(middleware.RequireRole("admin", "operator"))
				}
				{
					svcsWrite.POST("/:id/restart", servicesH.Restart)
					svcsWrite.POST("/:id/start", servicesH.Start)
					svcsWrite.POST("/:id/stop", servicesH.Stop)
					svcsWrite.POST("/:id/scale", servicesH.Scale)
				}
			}

			// Monitoring & Metrics (read-only for most roles)
			mon := apiProtected.Group("/monitoring")
			{
				mon.GET("/metrics", monitoringH.Metrics)
				mon.GET("/alerts", monitoringH.Alerts)
				mon.GET("/health", monitoringH.Health)
				mon.GET("/logs", monitoringH.Logs)
			}

			// Deployment Management
			deploy := apiProtected.Group("/deploy")
			{
				deploy.GET("/status", deploymentsH.Status)
				deploy.GET("/history", deploymentsH.History)

				// Write operations require higher permissions
				deployWrite := deploy.Group("/")
				if casbinSvc != nil {
					deployWrite.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/deploy", "POST"))
				} else {
					deployWrite.Use(middleware.RequireRole("admin", "operator"))
				}
				{
					deployWrite.POST("/start", deploymentsH.Start)
					deployWrite.POST("/rollback", deploymentsH.Rollback)
				}
			}

			// Plugin Management
			plugs := apiProtected.Group("/plugins")
			{
				plugs.GET("", pluginsH.List)

				// Write operations require admin permissions
				plugsWrite := plugs.Group("/")
				if casbinSvc != nil {
					plugsWrite.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/plugins", "POST"))
					plugsDelete := plugs.Group("/")
					plugsDelete.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/plugins", "DELETE"))
					plugsDelete.DELETE("/:id", pluginsH.Uninstall)
				} else {
					plugsWrite.Use(middleware.RequireRole("admin"))
					plugsWrite.DELETE("/:id", pluginsH.Uninstall)
				}
				{
					plugsWrite.POST("/install", pluginsH.Install)
					plugsWrite.POST("/:id/enable", pluginsH.Enable)
					plugsWrite.POST("/:id/disable", pluginsH.Disable)
				}
			}

			// Automation Management
			auto := apiProtected.Group("/automation")
			{
				auto.GET("", automationH.List)
				auto.GET("/logs", automationH.Logs)

				// Write operations require higher permissions
				autoWrite := auto.Group("/")
				if casbinSvc != nil {
					autoWrite.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/automation", "POST"))
				} else {
					autoWrite.Use(middleware.RequireRole("admin", "operator"))
				}
				{
					autoWrite.POST("", automationH.Create)
					autoWrite.POST("/:id/run", automationH.Run)
					autoWrite.POST("/:id/schedule", automationH.Schedule)
				}
			}

			// Billing & Invoicing
			bill := apiProtected.Group("/billing")
			{
				bill.GET("/invoices", billingH.ListInvoices)
				bill.GET("/payments", billingH.ListPayments)

				// Invoice generation requires admin permissions
				billWrite := bill.Group("/")
				if casbinSvc != nil {
					billWrite.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/billing", "POST"))
				} else {
					billWrite.Use(middleware.RequireRole("admin"))
				}
				{
					billWrite.POST("/invoices", billingH.GenerateInvoice)
				}
			}

			// Configuration Management
			cfg := apiProtected.Group("/config")
			{
				cfg.GET("", configH.Get)
				cfg.GET("/validate", configH.Validate)

				// Configuration updates require admin permissions
				cfgWrite := cfg.Group("/")
				if casbinSvc != nil {
					cfgWrite.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/config", "POST"))
				} else {
					cfgWrite.Use(middleware.RequireRole("admin"))
				}
				{
					cfgWrite.POST("", configH.Update)
				}
			}

			// Chaos Engineering
			chaosGroup := apiProtected.Group("/chaos")
			{
				chaosGroup.GET("/experiments", chaosH.List)
				chaosGroup.GET("/status", chaosH.Status)

				// Running chaos experiments requires admin permissions
				chaosWrite := chaosGroup.Group("/")
				if casbinSvc != nil {
					chaosWrite.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/chaos", "POST"))
				} else {
					chaosWrite.Use(middleware.RequireRole("admin"))
				}
				{
					chaosWrite.POST("/experiments", chaosH.Run)
				}
			}
		}
	}

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start metrics server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		metricsAddr := cfg.Server.MetricsPort
		if metricsAddr == "" {
			metricsAddr = "9090"
		}
		log.Printf("Starting metrics server on port %s", metricsAddr)
		if err := prometheusCollector.StartMetricsServer(ctx, ":"+metricsAddr); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Start periodic metrics collection
	go startMetricsCollection(ctx, metricsCollector, db)

	// Start alert manager
	go monitoring.StartAlertManager(ctx)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting API server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// startMetricsCollection periodically collects and stores metrics
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

// collectMetrics gathers metrics from various sources
func collectMetrics(ctx context.Context, mc *metrics.MetricsCollectorWithTimeSeries, db *database.Database) {
	// Collect subscriber metrics
	totalSubs, activeSubs, suspendedSubs := getSubscriberStats(db)
	mc.UpdateSubscriberMetrics(totalSubs, activeSubs, suspendedSubs)

	// Collect system metrics
	systemMetrics := metrics.SystemMetrics{
		ActiveSessions: getActiveSessionsCount(db),
		CPUUsage:       getCPUUsage(),
		MemoryUsage:    getMemoryUsage(),
		NetworkRX:      getNetworkRX(),
		NetworkTX:      getNetworkTX(),
		Timestamp:      time.Now(),
	}

	// Store in time-series if available
	if ts := mc.GetTimeSeriesStorage(); ts != nil {
		if err := ts.StoreSystemMetrics(ctx, systemMetrics); err != nil {
			log.Printf("Failed to store system metrics: %v", err)
		}
	}
}

// Helper functions for metrics collection (actual implementations)
func getSubscriberStats(db *database.Database) (total, active, suspended int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count total subscribers using GORM
	var totalCount int64
	if err := db.DB.WithContext(ctx).Model(&models.Subscriber{}).Count(&totalCount).Error; err != nil {
		log.Printf("Error counting total subscribers: %v", err)
		return 0, 0, 0
	}

	// Count active subscribers
	var activeCount int64
	if err := db.DB.WithContext(ctx).Model(&models.Subscriber{}).Where("status = ?", "active").Count(&activeCount).Error; err != nil {
		log.Printf("Error counting active subscribers: %v", err)
		return int(totalCount), 0, 0
	}

	// Count suspended subscribers
	var suspendedCount int64
	if err := db.DB.WithContext(ctx).Model(&models.Subscriber{}).Where("status IN ?", []string{"suspended", "deactivated", "blocked"}).Count(&suspendedCount).Error; err != nil {
		log.Printf("Error counting suspended subscribers: %v", err)
		return int(totalCount), int(activeCount), 0
	}

	return int(totalCount), int(activeCount), int(suspendedCount)
}

func getActiveSessionsCount(db *database.Database) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count active charging sessions using GORM
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

func getCPUUsage() float64 {
	// Read CPU usage from /proc/stat on Linux systems
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		log.Printf("Error reading /proc/stat: %v", err)
		return 0
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0
	}

	// Parse the first line (total CPU usage)
	fields := strings.Fields(lines[0])
	if len(fields) < 8 || fields[0] != "cpu" {
		return 0
	}

	// Parse CPU time values (user, nice, system, idle, iowait, irq, softirq, steal)
	var totalCPU, idleCPU float64
	for i, field := range fields[1:8] {
		val, err := strconv.ParseFloat(field, 64)
		if err != nil {
			continue
		}
		totalCPU += val
		if i == 3 { // idle time
			idleCPU = val
		}
	}

	if totalCPU == 0 {
		return 0
	}

	// Calculate CPU usage percentage (non-idle time)
	cpuUsage := ((totalCPU - idleCPU) / totalCPU) * 100
	return math.Min(cpuUsage, 100.0) // Cap at 100%
}

func getMemoryUsage() float64 {
	// Read memory usage from /proc/meminfo on Linux systems
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		log.Printf("Error reading /proc/meminfo: %v", err)
		return 0
	}

	lines := strings.Split(string(data), "\n")
	var totalMem, availableMem float64

	// Parse memory information
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				totalMem = val * 1024 // Convert KB to bytes
			}
		case "MemAvailable:":
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				availableMem = val * 1024 // Convert KB to bytes
			}
		}
	}

	if totalMem == 0 {
		return 0
	}

	// Calculate memory usage percentage
	usedMem := totalMem - availableMem
	memoryUsage := (usedMem / totalMem) * 100
	return math.Min(memoryUsage, 100.0) // Cap at 100%
}

func getNetworkRX() int64 {
	// Read network statistics from /proc/net/dev on Linux systems
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		log.Printf("Error reading /proc/net/dev: %v", err)
		return 0
	}

	lines := strings.Split(string(data), "\n")
	var totalRX int64

	// Skip header lines (first two lines)
	for i, line := range lines {
		if i < 2 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// Skip loopback interface
		if strings.HasPrefix(fields[0], "lo:") {
			continue
		}

		// RX bytes are in field 1 (after interface name)
		if rxBytes, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
			totalRX += rxBytes
		}
	}

	return totalRX
}

func getNetworkTX() int64 {
	// Read network statistics from /proc/net/dev on Linux systems
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		log.Printf("Error reading /proc/net/dev: %v", err)
		return 0
	}

	lines := strings.Split(string(data), "\n")
	var totalTX int64

	// Skip header lines (first two lines)
	for i, line := range lines {
		if i < 2 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// Skip loopback interface
		if strings.HasPrefix(fields[0], "lo:") {
			continue
		}

		// TX bytes are in field 9 (after interface name and RX stats)
		if txBytes, err := strconv.ParseInt(fields[9], 10, 64); err == nil {
			totalTX += txBytes
		}
	}

	return totalTX
}
