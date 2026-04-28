package payment

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListPaymentMethods lists payment methods for the authenticated subscriber.
func (pc *PaymentController) ListPaymentMethods(c *gin.Context) {
	subscriberID := c.GetInt("subscriberID")
	if subscriberID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	paymentMethods, err := pc.db.ListPaymentMethods(c.Request.Context(), uint(subscriberID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"paymentMethods": paymentMethods})
}

// ListTransactions lists transactions for the authenticated subscriber.
func (pc *PaymentController) ListTransactions(c *gin.Context) {
	subscriberID := c.GetInt("subscriberID")
	if subscriberID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	transactions, err := pc.db.ListTransactions(c.Request.Context(), uint(subscriberID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// GetTransaction returns a specific transaction, enforcing ownership.
func (pc *PaymentController) GetTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID required"})
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
	c.JSON(http.StatusOK, gin.H{"transaction": transaction})
}
