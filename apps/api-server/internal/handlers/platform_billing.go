package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// BillingHandler exposes invoice and payment endpoints for platform admins.
type BillingHandler struct {
	invoice *services.InvoiceService
	db      *gorm.DB
}

// NewBillingHandler constructs a BillingHandler.
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
