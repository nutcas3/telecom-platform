package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
)

// CurrencyHandler handles currency-related HTTP requests
type CurrencyHandler struct {
	billingService  currency.BillingService
	exchangeService currency.ExchangeRateService
	logger          *logrus.Logger
}

// NewCurrencyHandler creates a new currency handler
func NewCurrencyHandler(billingService currency.BillingService, exchangeService currency.ExchangeRateService, logger *logrus.Logger) *CurrencyHandler {
	return &CurrencyHandler{
		billingService:  billingService,
		exchangeService: exchangeService,
		logger:          logger,
	}
}

// ConvertCurrency handles currency conversion requests
func (h *CurrencyHandler) ConvertCurrency(c *gin.Context) {
	var req currency.CurrencyConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid currency conversion request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.billingService.ConvertAmount(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Currency conversion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetExchangeRate handles exchange rate requests
func (h *CurrencyHandler) GetExchangeRate(c *gin.Context) {
	fromCurrency := c.Param("from")
	toCurrency := c.Param("to")

	if fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to currency parameters are required"})
		return
	}

	rate, err := h.exchangeService.GetExchangeRate(c.Request.Context(), fromCurrency, toCurrency)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get exchange rate")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rate)
}

// ProcessBilling handles billing requests
func (h *CurrencyHandler) ProcessBilling(c *gin.Context) {
	var req currency.BillingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid billing request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.billingService.ProcessBilling(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Billing processing failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBillingHistory handles billing history requests
func (h *CurrencyHandler) GetBillingHistory(c *gin.Context) {
	profileID := c.Param("profile_id")
	if profileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile_id parameter is required"})
		return
	}

	// Parse query parameters
	filter := &currency.TransactionFilter{}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}
	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			filter.FromDate = &fromDate
		}
	}
	if toDateStr := c.Query("to_date"); toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			filter.ToDate = &toDate
		}
	}

	transactions, err := h.billingService.GetBillingHistory(c.Request.Context(), profileID, filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get billing history")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"count":        len(transactions),
	})
}

// GetBillingSummary handles billing summary requests
func (h *CurrencyHandler) GetBillingSummary(c *gin.Context) {
	profileID := c.Param("profile_id")
	if profileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile_id parameter is required"})
		return
	}

	// Parse date parameters
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	var fromDate, toDate time.Time
	var err error

	if fromDateStr != "" {
		fromDate, err = time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format, use YYYY-MM-DD"})
			return
		}
	} else {
		fromDate = time.Now().AddDate(0, -1, 0) // Default to 1 month ago
	}

	if toDateStr != "" {
		toDate, err = time.Parse("2006-01-02", toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format, use YYYY-MM-DD"})
			return
		}
	} else {
		toDate = time.Now() // Default to now
	}

	summary, err := h.billingService.CalculateTotalBilling(c.Request.Context(), profileID, fromDate, toDate)
	if err != nil {
		h.logger.WithError(err).Error("Failed to calculate billing summary")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ProcessRefund handles refund requests
func (h *CurrencyHandler) ProcessRefund(c *gin.Context) {
	transactionID := c.Param("transaction_id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction_id parameter is required"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,min=0"`
		Reason string  `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid refund request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process refund as a negative billing entry
	refundReq := &currency.BillingRequest{
		ProfileID:   c.Param("profile_id"),
		Amount:      -req.Amount, // Negative amount for refund
		Currency:    "USD",
		Description: "Refund: " + req.Reason,
		BillingDate: time.Now(),
	}

	if refundReq.ProfileID == "" {
		refundReq.ProfileID = transactionID // Use transaction ID as fallback reference
	}

	resp, err := h.billingService.ProcessBilling(c.Request.Context(), refundReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process refund")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refund_id":      resp.TransactionID,
		"transaction_id": transactionID,
		"amount":         req.Amount,
		"reason":         req.Reason,
		"status":         resp.Status,
		"processed_at":   resp.ProcessedAt,
	})
}

// GetBillingAnalytics handles billing analytics requests
func (h *CurrencyHandler) GetBillingAnalytics(c *gin.Context) {
	profileID := c.Query("profile_id")
	if profileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile_id query parameter is required"})
		return
	}

	// Default to last 30 days
	toDate := time.Now()
	fromDate := toDate.AddDate(0, -1, 0)

	if from := c.Query("from"); from != "" {
		if parsed, err := time.Parse(time.RFC3339, from); err == nil {
			fromDate = parsed
		}
	}
	if to := c.Query("to"); to != "" {
		if parsed, err := time.Parse(time.RFC3339, to); err == nil {
			toDate = parsed
		}
	}

	summary, err := h.billingService.CalculateTotalBilling(c.Request.Context(), profileID, fromDate, toDate)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get billing analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	history, err := h.billingService.GetBillingHistory(c.Request.Context(), profileID, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get billing history for analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary":           summary,
		"transaction_count": len(history),
		"from_date":         fromDate,
		"to_date":           toDate,
	})
}

// GetSupportedCurrencies handles supported currencies requests
func (h *CurrencyHandler) GetSupportedCurrencies(c *gin.Context) {
	currencies, err := h.exchangeService.GetSupportedCurrencies(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get supported currencies")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currencies": currencies,
		"count":      len(currencies),
	})
}

// RefreshExchangeRates handles exchange rate refresh requests
func (h *CurrencyHandler) RefreshExchangeRates(c *gin.Context) {
	err := h.exchangeService.RefreshRates(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to refresh exchange rates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exchange rates refreshed successfully"})
}

// GetExchangeRateHistory handles exchange rate history requests
func (h *CurrencyHandler) GetExchangeRateHistory(c *gin.Context) {
	fromCurrency := c.Param("from")
	toCurrency := c.Param("to")
	daysStr := c.Query("days")

	if fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to currency parameters are required"})
		return
	}

	days := 30 // Default to 30 days
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	rates, err := h.exchangeService.GetRateHistory(c.Request.Context(), fromCurrency, toCurrency, days)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get exchange rate history")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rates": rates,
		"count": len(rates),
		"from":  fromCurrency,
		"to":    toCurrency,
		"days":  days,
	})
}
