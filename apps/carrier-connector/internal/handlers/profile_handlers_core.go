package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/es2"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/mq"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/webhook"
)

// OrderProfileHandlerWithRepo orders a profile via ES2+ and persists it in the repo.
func OrderProfileHandlerWithRepo(client *es2.ES2Client, repo repository.ProfileRepository, webhookClient *webhook.WebhookClient, messageQueue *mq.MessageQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		var order ProfileOrder
		if err := c.ShouldBindJSON(&order); err != nil {
			Logger.Error().Err(err).Msg("Invalid request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "message": err.Error()})
			return
		}

		if order.ICCID == "" || order.IMSI == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Missing required fields",
				"message": "ICCID and IMSI are required",
			})
			return
		}

		Logger.Info().
			Str("imsi", order.IMSI).
			Str("iccid", order.ICCID).
			Msg("Ordering eSIM profile from SM-DP+")

		_, err := client.DownloadProfile(context.Background(), &es2.DownloadProfileRequest{
			EID:              order.EID,
			ICCID:            order.ICCID,
			ProfileType:      order.ProfileType,
			ConfirmationCode: order.ConfirmationCode,
		})
		if err != nil {
			Logger.Error().Err(err).Str("imsi", order.IMSI).Msg("Failed to order profile")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to download profile",
				"message": err.Error(),
			})
			return
		}

		// Persist the downloaded profile using correct struct fields
		profile := &repository.Profile{
			ICCID:       order.ICCID,
			EID:         order.EID,
			IMSI:        order.IMSI,
			MCC:         order.MCC,
			MNC:         order.MNC,
			ProfileType: order.ProfileType,
			State:       "downloaded",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Use Create method from repository interface
		if err := repo.Create(context.Background(), profile); err != nil {
			Logger.Error().Err(err).Str("imsi", order.IMSI).Msg("Failed to persist profile")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to persist profile",
				"message": err.Error(),
			})
			return
		}

		Logger.Info().
			Str("imsi", order.IMSI).
			Str("iccid", order.ICCID).
			Msg("Successfully ordered and persisted eSIM profile")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Profile ordered successfully",
			"profile": gin.H{
				"iccid":        order.ICCID,
				"imsi":         order.IMSI,
				"profile_type": order.ProfileType,
				"state":        "downloaded",
				"created_at":   time.Now().Format(time.RFC3339),
			},
		})
	}
}
