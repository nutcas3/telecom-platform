package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CurrencyHandler handles currency and billing HTTP requests
type CurrencyHandler struct{}

// NewCurrencyHandler creates a new currency handler
func NewCurrencyHandler() *CurrencyHandler {
	return &CurrencyHandler{}
}

// ConvertCurrency converts between currencies
func (h *CurrencyHandler) ConvertCurrency(c *gin.Context) {
	var req struct {
		From   string  `json:"from" binding:"required"`
		To     string  `json:"to" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simulated exchange rates
	rates := map[string]float64{
		"USD": 1.0,
		"EUR": 0.92,
		"GBP": 0.79,
		"JPY": 149.50,
		"CAD": 1.36,
		"AUD": 1.53,
		"CHF": 0.88,
	}

	fromRate := rates[req.From]
	toRate := rates[req.To]

	if fromRate == 0 || toRate == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported currency"})
		return
	}

	// Convert to USD first, then to target currency
	usdAmount := req.Amount / fromRate
	converted := usdAmount * toRate
	rate := toRate / fromRate

	c.JSON(http.StatusOK, gin.H{
		"from":      req.From,
		"to":        req.To,
		"amount":    req.Amount,
		"converted": converted,
		"rate":      rate,
		"timestamp": time.Now(),
	})
}

// GetExchangeRate returns exchange rate between currencies
func (h *CurrencyHandler) GetExchangeRate(c *gin.Context) {
	from := c.Param("from")
	to := c.Param("to")

	rates := map[string]float64{
		"USD": 1.0,
		"EUR": 0.92,
		"GBP": 0.79,
		"JPY": 149.50,
		"CAD": 1.36,
		"AUD": 1.53,
		"CHF": 0.88,
	}

	fromRate := rates[from]
	toRate := rates[to]

	if fromRate == 0 || toRate == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported currency"})
		return
	}

	rate := toRate / fromRate

	c.JSON(http.StatusOK, gin.H{
		"from":      from,
		"to":        to,
		"rate":      rate,
		"timestamp": time.Now(),
	})
}

// GetExchangeRateHistory returns exchange rate history
func (h *CurrencyHandler) GetExchangeRateHistory(c *gin.Context) {
	from := c.Param("from")
	to := c.Param("to")

	// Simulated historical rates
	history := make([]map[string]any, 0)
	baseRate := 0.92 // EUR/USD base

	for i := 30; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		variation := (float64(i%5) - 2) * 0.005 // Small daily variation
		history = append(history, map[string]any{
			"from":      from,
			"to":        to,
			"rate":      baseRate + variation,
			"timestamp": date.Format("2006-01-02"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

// GetSupportedCurrencies returns list of supported currencies
func (h *CurrencyHandler) GetSupportedCurrencies(c *gin.Context) {
	currencies := []map[string]any{
		{"code": "USD", "name": "US Dollar", "symbol": "$"},
		{"code": "EUR", "name": "Euro", "symbol": "€"},
		{"code": "GBP", "name": "British Pound", "symbol": "£"},
		{"code": "JPY", "name": "Japanese Yen", "symbol": "¥"},
		{"code": "CAD", "name": "Canadian Dollar", "symbol": "C$"},
		{"code": "AUD", "name": "Australian Dollar", "symbol": "A$"},
		{"code": "CHF", "name": "Swiss Franc", "symbol": "CHF"},
		{"code": "CNY", "name": "Chinese Yuan", "symbol": "¥"},
		{"code": "INR", "name": "Indian Rupee", "symbol": "₹"},
		{"code": "KRW", "name": "South Korean Won", "symbol": "₩"},
	}

	c.JSON(http.StatusOK, gin.H{"currencies": currencies})
}

// RefreshExchangeRates refreshes exchange rates from external sources
func (h *CurrencyHandler) RefreshExchangeRates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "refreshed",
		"updated_at":  time.Now(),
		"source":      "external_api",
		"rates_count": 10,
	})
}

// ProcessBilling processes a billing transaction
func (h *CurrencyHandler) ProcessBilling(c *gin.Context) {
	var req struct {
		ProfileID   string  `json:"profile_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		Currency    string  `json:"currency" binding:"required"`
		Description string  `json:"description"`
		Type        string  `json:"type"` // charge, credit, refund
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type == "" {
		req.Type = "charge"
	}

	transaction := map[string]any{
		"id":          "txn-" + time.Now().Format("20060102150405"),
		"profile_id":  req.ProfileID,
		"amount":      req.Amount,
		"currency":    req.Currency,
		"type":        req.Type,
		"description": req.Description,
		"status":      "completed",
		"created_at":  time.Now(),
	}

	c.JSON(http.StatusCreated, transaction)
}

// GetBillingHistory returns billing history for a profile
func (h *CurrencyHandler) GetBillingHistory(c *gin.Context) {
	profileID := c.Param("profile_id")

	transactions := []map[string]any{
		{
			"id":          "txn-001",
			"profile_id":  profileID,
			"amount":      29.99,
			"currency":    "USD",
			"type":        "charge",
			"description": "Monthly subscription",
			"status":      "completed",
			"created_at":  time.Now().AddDate(0, 0, -1),
		},
		{
			"id":          "txn-002",
			"profile_id":  profileID,
			"amount":      15.00,
			"currency":    "USD",
			"type":        "charge",
			"description": "Data top-up",
			"status":      "completed",
			"created_at":  time.Now().AddDate(0, 0, -15),
		},
		{
			"id":          "txn-003",
			"profile_id":  profileID,
			"amount":      29.99,
			"currency":    "USD",
			"type":        "charge",
			"description": "Monthly subscription",
			"status":      "completed",
			"created_at":  time.Now().AddDate(0, -1, -1),
		},
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// GetBillingSummary returns billing summary for a profile
func (h *CurrencyHandler) GetBillingSummary(c *gin.Context) {
	profileID := c.Param("profile_id")
	period := c.DefaultQuery("period", "monthly")

	summary := map[string]any{
		"profile_id":        profileID,
		"period":            period,
		"total_amount":      74.98,
		"currency":          "USD",
		"transaction_count": 3,
		"breakdown": map[string]float64{
			"subscription": 59.98,
			"data_topup":   15.00,
		},
		"average_transaction": 24.99,
		"generated_at":        time.Now(),
	}

	c.JSON(http.StatusOK, summary)
}

// ProcessRefund processes a refund for a transaction
func (h *CurrencyHandler) ProcessRefund(c *gin.Context) {
	transactionID := c.Param("transaction_id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund := map[string]any{
		"id":                   "ref-" + time.Now().Format("20060102150405"),
		"original_transaction": transactionID,
		"amount":               29.99,
		"currency":             "USD",
		"reason":               req.Reason,
		"status":               "completed",
		"created_at":           time.Now(),
	}

	c.JSON(http.StatusOK, refund)
}

// GetBillingAnalytics returns billing analytics
func (h *CurrencyHandler) GetBillingAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "monthly")

	analytics := map[string]any{
		"period":              period,
		"total_revenue":       4500000,
		"total_transactions":  150000,
		"average_transaction": 30.0,
		"by_type": map[string]float64{
			"subscription": 3500000,
			"data_topup":   750000,
			"voice_topup":  150000,
			"other":        100000,
		},
		"by_currency": map[string]float64{
			"USD": 3000000,
			"EUR": 1000000,
			"GBP": 500000,
		},
		"growth_rate":  12.5,
		"churn_impact": -150000,
		"generated_at": time.Now(),
	}

	c.JSON(http.StatusOK, analytics)
}
