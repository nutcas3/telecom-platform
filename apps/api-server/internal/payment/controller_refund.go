package payment

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// RefundTransaction refunds a completed transaction, creating a reverse
// Transaction record and adjusting subscriber balance.
func (pc *PaymentController) RefundTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID required"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,min=0.01"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscriberID := c.GetInt("subscriberID")
	if subscriberID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	transaction, err := pc.db.GetTransaction(c.Request.Context(), transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	if transaction.SubscriberID != uint(subscriberID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Transaction does not belong to subscriber"})
		return
	}
	if transaction.Status != "COMPLETED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot refund non-completed transaction"})
		return
	}

	refundResp, err := pc.paymentService.gateway.RefundPayment(c.Request.Context(), transaction.TransactionID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refund := &models.Transaction{
		SubscriberID:  transaction.SubscriberID,
		TransactionID: refundResp.RefundID,
		Type:          "REFUND",
		Amount:        req.Amount,
		Currency:      refundResp.Currency,
		Status:        string(refundResp.Status),
		Description:   fmt.Sprintf("Refund for transaction %s", transactionID),
		ParentID:      &transaction.ID,
		CreatedAt:     refundResp.ProcessedAt,
		UpdatedAt:     refundResp.ProcessedAt,
	}

	if err := pc.db.CreateTransaction(c.Request.Context(), refund); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refund"})
		return
	}

	if err := pc.db.UpdateSubscriberBalance(c.Request.Context(), transaction.SubscriberID, -req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refund":  refund,
		"message": "Refund processed successfully",
	})
}

// HandleStripeWebhook delegates to WebhookHandler.
func (pc *PaymentController) HandleStripeWebhook(c *gin.Context) {
	NewWebhookHandler(pc.paymentService, pc.db).HandleStripeWebhook(c)
}
