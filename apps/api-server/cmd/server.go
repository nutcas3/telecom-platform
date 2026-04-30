package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	csrf "github.com/utrack/gin-csrf"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/discovery"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/middleware"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/monitoring"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/websocket"
)

// @title Telecom Platform API
// @version 1.0
// @description API for Telecom Platform
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.LoadConfig()

	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		return
	}
	defer db.Close()

	// Initialize service discovery (Consul)
	var serviceDiscovery *discovery.ServiceDiscovery
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr != "" {
		sd, err := discovery.NewServiceDiscovery()
		if err != nil {
			log.Printf("Failed to initialize service discovery: %v", err)
		} else {
			serviceDiscovery = sd
			// Register this service with Consul
			port, err := strconv.Atoi(cfg.Server.Port)
			if err != nil {
				log.Printf("Failed to parse port: %v", err)
				port = 8080 // fallback
			}
			serviceID := fmt.Sprintf("api-server-%s", cfg.Server.Port)
			err = serviceDiscovery.Register(discovery.Service{
				ID:      serviceID,
				Name:    "api-server",
				Address: "0.0.0.0",
				Port:    port,
				Tags:    []string{"api", "telecom"},
			}, 10*time.Second)
			if err != nil {
				log.Printf("Failed to register service with Consul: %v", err)
			} else {
				log.Printf("Service registered with Consul: %s", serviceID)
				defer serviceDiscovery.Deregister(serviceID)
			}
		}
	}

	prometheusCollector, metricsCollector, timeSeriesStorage := buildMetricsCollector(cfg)
	defer func() {
		if timeSeriesStorage != nil {
			timeSeriesStorage.Close()
		}
	}()

	router := gin.Default()

	// Add CSRF protection middleware
	csrfSecret := os.Getenv("CSRF_SECRET")
	if csrfSecret == "" {
		csrfSecret = "change-me-in-production-csrf-secret-32-chars-min"
		log.Printf("WARNING: Using default CSRF secret, set CSRF_SECRET environment variable")
	}
	router.Use(csrf.Middleware(csrf.Options{
		Secret: csrfSecret,
	}))

	router.Use(func(c *gin.Context) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { c.Next() })
		prometheusCollector.HTTPMiddleware(next).ServeHTTP(c.Writer, c.Request)
	})

	router.GET("/metrics", gin.WrapH(prometheusCollector.Handler()))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/ws", websocket.HandleWebSocket(websocket.GetHub()))

	router.GET("/health", monitoring.HealthHandler())
	router.GET("/ready", monitoring.ReadyHandler())
	router.GET("/live", monitoring.LiveHandler())

	router.Use(middleware.RateLimitByEndpoint())
	router.Use(middleware.PerformanceMiddlewareHandler())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestSizeMiddleware(10 << 20))
	router.Use(middleware.TimeoutMiddleware(30 * time.Second))
	router.GET("/metrics/performance", middleware.PerformanceHandler())

	websocket.InitializeWebSocket()
	monitoring.InitializeHealthMonitor("1.0.0", "production")
	healthMonitor := monitoring.GetHealthMonitor()
	healthMonitor.RegisterCheck("database", monitoring.NewDatabaseHealthChecker(db.DB))
	healthMonitor.RegisterCheck("system", monitoring.NewSystemHealthChecker())
	monitoring.InitializeAlertManager(healthMonitor)
	middleware.InitializePerformanceMiddleware(100 * time.Millisecond)

	deps := buildDeps(db, cfg)
	registerV1Routes(router, deps)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

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

	go startMetricsCollection(ctx, metricsCollector, db)
	go monitoring.StartAlertManager(ctx)

	go func() {
		log.Printf("Starting API server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
