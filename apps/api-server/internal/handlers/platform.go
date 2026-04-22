package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/monitoring"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// ErrorResponse is a shared error envelope used across platform handlers.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Details: err.Error(), Code: "BAD_REQUEST"})
}

func serverError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal error", Details: err.Error(), Code: "INTERNAL_ERROR"})
}

func notFound(c *gin.Context, err error) {
	c.JSON(http.StatusNotFound, ErrorResponse{Error: "Not found", Details: err.Error(), Code: "NOT_FOUND"})
}

// parseUint helper
func parseUintParam(c *gin.Context, name string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		badRequest(c, err)
		return 0, false
	}
	return uint(v), true
}

// parseBoolQuery returns *bool for optional tri-state filters.
func parseBoolQuery(c *gin.Context, key string) *bool {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &b
}

type ServicesHandler struct {
	k8s *services.KubernetesService
}

func NewServicesHandler(k8s *services.KubernetesService) *ServicesHandler {
	return &ServicesHandler{k8s: k8s}
}

func (h *ServicesHandler) List(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	deps, err := h.k8s.ListDeployments(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"services": deps, "total": len(deps), "namespace": h.k8s.Namespace()})
}

func (h *ServicesHandler) Get(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	d, err := h.k8s.GetDeployment(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFound(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *ServicesHandler) Restart(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	if err := h.k8s.RestartDeployment(c.Request.Context(), c.Param("id")); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "restart triggered", "service": c.Param("id")})
}

func (h *ServicesHandler) Start(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	// Starting = scaling to >=1 replica (default 1).
	if err := h.k8s.ScaleDeployment(c.Request.Context(), c.Param("id"), 1); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "service started", "service": c.Param("id")})
}

func (h *ServicesHandler) Stop(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	if err := h.k8s.ScaleDeployment(c.Request.Context(), c.Param("id"), 0); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "service stopped", "service": c.Param("id")})
}

func (h *ServicesHandler) Scale(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	var req struct {
		Replicas int32 `json:"replicas" binding:"required,min=0,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.k8s.ScaleDeployment(c.Request.Context(), c.Param("id"), req.Replicas); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "scaled", "service": c.Param("id"), "replicas": req.Replicas})
}

func (h *ServicesHandler) Logs(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	lines, _ := strconv.Atoi(c.DefaultQuery("lines", "100"))
	logs, err := h.k8s.PodLogs(c.Request.Context(), c.Param("id"), lines)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"service": c.Param("id"), "logs": logs})
}

func (h *ServicesHandler) Health(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	d, err := h.k8s.GetDeployment(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFound(c, err)
		return
	}
	status := "healthy"
	if d.ReadyReplicas < d.Replicas {
		status = "degraded"
	}
	if d.ReadyReplicas == 0 {
		status = "unhealthy"
	}
	c.JSON(http.StatusOK, gin.H{
		"service":            c.Param("id"),
		"status":             status,
		"replicas_desired":   d.Replicas,
		"replicas_ready":     d.ReadyReplicas,
		"replicas_available": d.AvailableReplicas,
	})
}

func (h *ServicesHandler) PodStatus(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	status, err := h.k8s.GetPodStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *ServicesHandler) Events(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	events, err := h.k8s.GetEvents(c.Request.Context(), c.Param("id"), limit)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events, "total": len(events)})
}

type MonitoringHandler struct {
	prom *monitoring.PrometheusService
}

func NewMonitoringHandler(prom *monitoring.PrometheusService) *MonitoringHandler {
	return &MonitoringHandler{prom: prom}
}

func (h *MonitoringHandler) Metrics(c *gin.Context) {
	if h.prom == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "monitoring service not available"})
		return
	}
	q := c.DefaultQuery("query", "up")
	samples, err := h.prom.Query(c.Request.Context(), q)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"query": q, "samples": samples})
}

func (h *MonitoringHandler) Alerts(c *gin.Context) {
	if h.prom == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "monitoring service not available"})
		return
	}
	alerts, err := h.prom.Alerts(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}

func (h *MonitoringHandler) Health(c *gin.Context) {
	if h.prom == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":               "degraded",
			"prometheus_available": false,
			"error":                "monitoring service not available",
		})
		return
	}
	available := h.prom.Available(c.Request.Context())
	status := "healthy"
	if !available {
		status = "degraded"
	}
	c.JSON(http.StatusOK, gin.H{
		"status":               status,
		"prometheus_available": available,
	})
}

func (h *MonitoringHandler) Logs(c *gin.Context) {
	// Logs are aggregated in Elasticsearch; expose a stub that points callers
	// to /v1/services/:id/logs for live pod logs.
	c.JSON(http.StatusOK, gin.H{
		"message": "Use /v1/services/:id/logs for live pod logs; aggregated logs are in Kibana",
	})
}

type DeploymentsHandler struct {
	svc *services.DeploymentService
}

func NewDeploymentsHandler(svc *services.DeploymentService) *DeploymentsHandler {
	return &DeploymentsHandler{svc: svc}
}

func (h *DeploymentsHandler) Status(c *gin.Context) {
	f := services.DeploymentFilter{
		Service:     c.Query("service"),
		Environment: c.Query("environment"),
		Status:      "in_progress",
		Limit:       50,
	}
	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deployments": items, "total": total})
}

func (h *DeploymentsHandler) Start(c *gin.Context) {
	var in services.StartDeploymentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	d, err := h.svc.Start(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *DeploymentsHandler) Rollback(c *gin.Context) {
	var in services.RollbackInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	d, err := h.svc.Rollback(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *DeploymentsHandler) History(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}
	f := services.DeploymentFilter{
		Service:     c.Query("service"),
		Environment: c.Query("environment"),
		Status:      c.Query("status"),
		Limit:       pageSize,
		Offset:      (page - 1) * pageSize,
	}
	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"deployments": items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
	})
}

type PluginsHandler struct {
	svc *services.PluginService
}

func NewPluginsHandler(svc *services.PluginService) *PluginsHandler {
	return &PluginsHandler{svc: svc}
}

func (h *PluginsHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), services.PluginFilter{
		Enabled:  parseBoolQuery(c, "enabled"),
		Type:     c.Query("type"),
		Category: c.Query("category"),
	})
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plugins": items, "total": len(items)})
}

func (h *PluginsHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	plugin, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		notFound(c, err)
		return
	}

	c.JSON(http.StatusOK, plugin)
}

func (h *PluginsHandler) Install(c *gin.Context) {
	var in services.InstallPluginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	p, err := h.svc.Install(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *PluginsHandler) Uninstall(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Uninstall(c.Request.Context(), id); err != nil {
		notFound(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin uninstalled", "id": id})
}

func (h *PluginsHandler) Enable(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.SetEnabled(c.Request.Context(), id, true)
	if err != nil {
		notFound(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PluginsHandler) Disable(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.SetEnabled(c.Request.Context(), id, false)
	if err != nil {
		notFound(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

type AutomationHandler struct {
	svc *services.AutomationService
}

func NewAutomationHandler(svc *services.AutomationService) *AutomationHandler {
	return &AutomationHandler{svc: svc}
}

func (h *AutomationHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), services.AutomationFilter{
		Enabled: parseBoolQuery(c, "enabled"),
		Type:    c.Query("type"),
	})
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"automations": items, "total": len(items)})
}

func (h *AutomationHandler) Create(c *gin.Context) {
	var in services.CreateAutomationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	a, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *AutomationHandler) Run(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	r, err := h.svc.Run(c.Request.Context(), id)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *AutomationHandler) Schedule(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var in services.ScheduleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	a, err := h.svc.UpdateSchedule(c.Request.Context(), id, in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AutomationHandler) Logs(c *gin.Context) {
	var automationID uint
	if raw := c.Query("automation_id"); raw != "" {
		v, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			badRequest(c, err)
			return
		}
		automationID = uint(v)
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	runs, err := h.svc.ListRuns(c.Request.Context(), automationID, limit)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"runs": runs, "total": len(runs)})
}

type ConfigHandler struct {
	svc *services.ConfigStoreService
}

func NewConfigHandler(svc *services.ConfigStoreService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

func (h *ConfigHandler) Get(c *gin.Context) {
	entries, err := h.svc.List(c.Request.Context(), c.Query("section"))
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries, "total": len(entries)})
}

func (h *ConfigHandler) Update(c *gin.Context) {
	var in services.UpsertConfigInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	entry, err := h.svc.Upsert(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *ConfigHandler) Validate(c *gin.Context) {
	res, err := h.svc.Validate(c.Request.Context(), c.Query("section"))
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

type ChaosHandler struct {
	svc *services.ChaosService
}

func NewChaosHandler(svc *services.ChaosService) *ChaosHandler {
	return &ChaosHandler{svc: svc}
}

type runExperimentRequest struct {
	Type        services.ExperimentType `json:"type" binding:"required"`
	Target      string                  `json:"target" binding:"required"`
	DurationMS  int64                   `json:"duration_ms"`
	Probability float64                 `json:"probability"`
	Amount      int                     `json:"amount"`
}

func (h *ChaosHandler) Run(c *gin.Context) {
	var req runExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	cfg := services.ExperimentConfig{
		Duration:    millisToDuration(req.DurationMS),
		Probability: req.Probability,
		Amount:      req.Amount,
		Target:      req.Target,
	}
	exp, err := h.svc.RunExperiment(c.Request.Context(), req.Type, cfg)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, exp)
}

func (h *ChaosHandler) List(c *gin.Context) {
	active := h.svc.GetActiveExperiments()
	history, _ := h.svc.GetExperimentHistory()
	c.JSON(http.StatusOK, gin.H{
		"active":  active,
		"history": history,
	})
}

func (h *ChaosHandler) Status(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		// No id: return summary of active experiments
		active := h.svc.GetActiveExperiments()
		c.JSON(http.StatusOK, gin.H{
			"active_count": len(active),
			"active":       active,
		})
		return
	}
	exp, ok := h.svc.GetExperimentStatus(id)
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "experiment not found", Code: "NOT_FOUND"})
		return
	}
	c.JSON(http.StatusOK, exp)
}

type BillingHandler struct {
	invoice *services.InvoiceService
	db      *gorm.DB
}

func NewBillingHandler(invoice *services.InvoiceService, db *gorm.DB) *BillingHandler {
	return &BillingHandler{invoice: invoice, db: db}
}

type generateInvoiceRequest struct {
	SubscriberID  uint   `json:"subscriber_id" binding:"required"`
	BillingPeriod string `json:"billing_period" binding:"required"` // e.g. "2024-01"
}

func (h *BillingHandler) GenerateInvoice(c *gin.Context) {
	var req generateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	inv, err := h.invoice.GenerateMonthlyInvoice(c.Request.Context(), req.SubscriberID, req.BillingPeriod)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, inv)
}

func (h *BillingHandler) ListInvoices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}
	q := h.db.WithContext(c.Request.Context()).Model(&models.Invoice{}).Preload("LineItems")
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if sid := c.Query("subscriber_id"); sid != "" {
		q = q.Where("subscriber_id = ?", sid)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		serverError(c, err)
		return
	}
	var items []models.Invoice
	if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error; err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoices": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *BillingHandler) ListPayments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}
	q := h.db.WithContext(c.Request.Context()).Model(&models.Transaction{})
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if sid := c.Query("subscriber_id"); sid != "" {
		q = q.Where("subscriber_id = ?", sid)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		serverError(c, err)
		return
	}
	var items []models.Transaction
	if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error; err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"payments": items, "total": total, "page": page, "page_size": pageSize})
}
