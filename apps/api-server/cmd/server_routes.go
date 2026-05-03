package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/handlers"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/middleware"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/rbac"
)

// registerV1Routes wires the /v1 API routes to the provided deps.
// RBAC enforcement uses Casbin when configured, falling back to role-based
// middleware otherwise.
func registerV1Routes(router *gin.Engine, d *serverDeps) {
	v1 := router.Group("/api/v1")

	// Health check
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger documentation
	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Charging engine integration (public for packet-gateway integration)
	charging := v1.Group("/charging")
	{
		charging.POST("/credit/:ip/check", d.chargingH.CheckCredit)
		charging.GET("/credit/:ip/balance", d.chargingH.GetBalance)
		charging.POST("/credit/:ip/add", d.chargingH.AddCredit)
		charging.POST("/credit/:ip/deduct", d.chargingH.DeductCredit)
	}

	auth := v1.Group("/auth")
	{
		auth.POST("/login", d.authH.Login)
		auth.POST("/register", d.authH.Register)
		auth.POST("/refresh", d.authH.RefreshToken)
	}

	protected := v1.Group("/")
	protected.Use(middleware.RateLimitByUser(100)) // 100 requests per minute per user
	protected.Use(middleware.AuthMiddleware(d.authSvc))
	{
		authProtected := protected.Group("/auth")
		{
			authProtected.POST("/logout", d.authH.Logout)
			authProtected.GET("/profile", d.authH.GetProfile)
			authProtected.POST("/change-password", d.authH.ChangePassword)
		}

		users := protected.Group("/users")
		applyRBAC(users, d.casbinSvc, "/api/v1/users", "GET", "admin")
		{
			users.GET("", d.authH.GetUsers)
			users.POST("", d.authH.CreateUser)
			users.PUT("/:id", d.authH.UpdateUser)
			users.DELETE("/:id", d.authH.DeleteUser)
		}
	}

	apiProtected := v1.Group("/")
	apiProtected.Use(middleware.RateLimitByUser(100)) // 100 requests per minute per user
	apiProtected.Use(middleware.AuthMiddleware(d.authSvc))
	{
		registerSubscriberRoutes(apiProtected, d)
		registerServicesRoutes(apiProtected, d)
		registerMonitoringRoutes(apiProtected, d)
		registerDeployRoutes(apiProtected, d)
		registerPluginRoutes(apiProtected, d)
		registerAutomationRoutes(apiProtected, d)
		registerBillingRoutes(apiProtected, d)
		registerConfigRoutes(apiProtected, d)
		registerChaosRoutes(apiProtected, d)
		registerAnalyticsRoutes(apiProtected)
		registerSecurityRoutes(apiProtected)
		registerCurrencyRoutes(apiProtected)
	}
}

// applyRBAC attaches Casbin or fallback role middleware to a group.
func applyRBAC(g *gin.RouterGroup, casbinSvc *rbac.CasbinService, path, method string, fallbackRoles ...string) {
	if casbinSvc != nil {
		g.Use(middleware.RequireCasbinPermission(casbinSvc, path, method))
	} else {
		g.Use(middleware.RequireRole(fallbackRoles...))
	}
}

func registerServicesRoutes(api *gin.RouterGroup, d *serverDeps) {
	svcs := api.Group("/services")
	svcs.GET("", d.servicesH.List)
	svcs.GET("/:id", d.servicesH.Get)
	svcs.GET("/:id/logs", d.servicesH.Logs)
	svcs.GET("/:id/health", d.servicesH.Health)
	svcs.GET("/:id/pods", d.servicesH.PodStatus)
	svcs.GET("/:id/events", d.servicesH.Events)

	w := svcs.Group("/")
	applyRBAC(w, d.casbinSvc, "/v1/services", "POST", "admin", "operator")
	w.POST("/:id/restart", d.servicesH.Restart)
	w.POST("/:id/start", d.servicesH.Start)
	w.POST("/:id/stop", d.servicesH.Stop)
	w.POST("/:id/scale", d.servicesH.Scale)
}

func registerMonitoringRoutes(api *gin.RouterGroup, d *serverDeps) {
	mon := api.Group("/monitoring")
	mon.GET("/metrics", d.monitoringH.Metrics)
	mon.GET("/alerts", d.monitoringH.Alerts)
	mon.GET("/health", d.monitoringH.Health)
	mon.GET("/logs", d.monitoringH.Logs)
}

func registerDeployRoutes(api *gin.RouterGroup, d *serverDeps) {
	deploy := api.Group("/deploy")
	deploy.GET("/status", d.deploymentsH.Status)
	deploy.GET("/history", d.deploymentsH.History)

	w := deploy.Group("/")
	applyRBAC(w, d.casbinSvc, "/v1/deploy", "POST", "admin", "operator")
	w.POST("/start", d.deploymentsH.Start)
	w.POST("/rollback", d.deploymentsH.Rollback)
}

func registerPluginRoutes(api *gin.RouterGroup, d *serverDeps) {
	plugs := api.Group("/plugins")
	plugs.GET("", d.pluginsH.List)

	w := plugs.Group("/")
	applyRBAC(w, d.casbinSvc, "/v1/plugins", "POST", "admin")
	w.POST("/install", d.pluginsH.Install)
	w.POST("/:id/enable", d.pluginsH.Enable)
	w.POST("/:id/disable", d.pluginsH.Disable)

	if d.casbinSvc != nil {
		del := plugs.Group("/")
		del.Use(middleware.RequireCasbinPermission(d.casbinSvc, "/v1/plugins", "DELETE"))
		del.DELETE("/:id", d.pluginsH.Uninstall)
	} else {
		w.DELETE("/:id", d.pluginsH.Uninstall)
	}
}

func registerAutomationRoutes(api *gin.RouterGroup, d *serverDeps) {
	auto := api.Group("/automation")
	auto.GET("", d.automationH.List)
	auto.GET("/logs", d.automationH.Logs)

	w := auto.Group("/")
	applyRBAC(w, d.casbinSvc, "/v1/automation", "POST", "admin", "operator")
	w.POST("", d.automationH.Create)
	w.POST("/:id/run", d.automationH.Run)
	w.POST("/:id/schedule", d.automationH.Schedule)
}

func registerBillingRoutes(api *gin.RouterGroup, d *serverDeps) {
	bill := api.Group("/billing")
	bill.GET("/invoices", d.billingH.ListInvoices)
	bill.GET("/payments", d.billingH.ListPayments)

	w := bill.Group("/")
	applyRBAC(w, d.casbinSvc, "/v1/billing", "POST", "admin")
	w.POST("/invoices", d.billingH.GenerateInvoice)
}

func registerConfigRoutes(api *gin.RouterGroup, d *serverDeps) {
	cfg := api.Group("/config")
	cfg.GET("", d.configH.Get)
	cfg.GET("/validate", d.configH.Validate)

	w := cfg.Group("/")
	applyRBAC(w, d.casbinSvc, "/v1/config", "POST", "admin")
	w.POST("", d.configH.Update)
}

func registerChaosRoutes(api *gin.RouterGroup, d *serverDeps) {
	chaosGroup := api.Group("/chaos")
	chaosGroup.GET("/experiments", d.chaosH.List)
	chaosGroup.GET("/status", d.chaosH.Status)

	w := chaosGroup.Group("/")
	applyRBAC(w, d.casbinSvc, "/v1/chaos", "POST", "admin")
	w.POST("/experiments", d.chaosH.Run)
}

func registerSubscriberRoutes(api *gin.RouterGroup, d *serverDeps) {
	subs := api.Group("/subscribers")
	subs.GET("", d.subscriberH.ListSubscribers)
	subs.GET("/:id", d.subscriberH.GetSubscriber)
	subs.GET("/imsi/:imsi", d.subscriberH.GetSubscriberByIMSI)

	w := subs.Group("/")
	applyRBAC(w, d.casbinSvc, "/v1/subscribers", "POST", "admin", "operator")
	w.POST("", d.subscriberH.CreateSubscriber)
	w.PUT("/:id", d.subscriberH.UpdateSubscriber)
	w.POST("/:id/suspend", d.subscriberH.SuspendSubscriber)
	w.POST("/:id/terminate", d.subscriberH.TerminateSubscriber)
}

func registerAnalyticsRoutes(api *gin.RouterGroup) {
	h := handlers.NewAnalyticsHandler()
	analytics := api.Group("/analytics")

	// Churn Analysis
	churn := analytics.Group("/churn")
	churn.POST("/predict", h.PredictChurn)
	churn.GET("/metrics", h.GetChurnMetrics)
	churn.GET("/at-risk", h.GetAtRiskCustomers)

	// Market Analytics
	market := analytics.Group("/market")
	market.GET("/metrics", h.GetMarketMetrics)
	market.GET("/competitors", h.GetCompetitors)
	market.GET("/opportunities", h.GetMarketOpportunities)

	// Predictive Maintenance
	maintenance := analytics.Group("/maintenance")
	maintenance.GET("/metrics", h.GetMaintenanceMetrics)
	maintenance.GET("/assets", h.GetAssetsHealth)
	maintenance.GET("/alerts", h.GetMaintenanceAlerts)
	maintenance.POST("/predict/:asset_id", h.PredictFailure)

	// Pricing Optimization
	pricing := analytics.Group("/pricing")
	pricing.GET("/metrics", h.GetPricingMetrics)
	pricing.POST("/optimize", h.OptimizePricing)
	pricing.GET("/elasticity", h.GetPriceElasticity)
}

func registerSecurityRoutes(api *gin.RouterGroup) {
	h := handlers.NewSecurityHandler()
	security := api.Group("/security")

	// Fraud Detection
	fraud := security.Group("/fraud")
	fraud.POST("/analyze", h.AnalyzeTransaction)
	fraud.POST("/alerts", h.GetFraudAlerts)
	fraud.PUT("/alerts/:id", h.UpdateAlertStatus)
	fraud.GET("/metrics", h.GetFraudMetrics)
	fraud.GET("/patterns", h.GetFraudPatterns)

	// SIM Swap Protection
	simswap := security.Group("/simswap")
	simswap.POST("/verify", h.VerifySIMSwap)
	simswap.GET("/history/:profile_id", h.GetSIMSwapHistory)
}

func registerCurrencyRoutes(api *gin.RouterGroup) {
	h := handlers.NewCurrencyHandler()
	currencyGroup := api.Group("/currency")

	currencyGroup.POST("/convert", h.ConvertCurrency)
	currencyGroup.GET("/exchange/:from/:to", h.GetExchangeRate)
	currencyGroup.GET("/exchange/:from/:to/history", h.GetExchangeRateHistory)
	currencyGroup.GET("/currencies", h.GetSupportedCurrencies)
	currencyGroup.POST("/exchange/refresh", h.RefreshExchangeRates)

	currencyGroup.POST("/billing", h.ProcessBilling)
	currencyGroup.GET("/billing/history/:profile_id", h.GetBillingHistory)
	currencyGroup.GET("/billing/summary/:profile_id", h.GetBillingSummary)
	currencyGroup.POST("/billing/refund/:transaction_id", h.ProcessRefund)
	currencyGroup.GET("/billing/analytics", h.GetBillingAnalytics)
}
