package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/es2"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/webhook"
)

// OrderProfileHandlerWithRepo orders a profile via ES2+ and persists it in the repo.
func OrderProfileHandlerWithRepo(client *es2.ES2Client, repo repository.ProfileRepository, webhookClient *webhook.WebhookClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var order ProfileOrder
		if err := c.ShouldBindJSON(&order); err != nil {
			Logger.Error().Err(err).Msg("Invalid request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "message": err.Error()})
			return
		}

		if order.ICCID == "" || order.IMSI == "" || order.K == "" || order.OPc == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Missing required fields",
				"message": "ICCID, IMSI, K, and OPc are required",
			})
			return
		}

		Logger.Info().
			Str("imsi", order.IMSI).
			Str("iccid", order.ICCID).
			Msg("Ordering eSIM profile from SM-DP+")

		downloadResp, err := client.DownloadProfile(context.Background(), &es2.DownloadProfileRequest{
			EID:              order.EID,
			ICCID:            order.ICCID,
			ProfileType:      order.ProfileType,
			ConfirmationCode: order.ConfirmationCode,
		})
		if err != nil {
			Logger.Error().Err(err).Str("imsi", order.IMSI).Msg("Failed to order profile")

			// Send webhook notification for failed download
			if webhookClient != nil {
				go webhookClient.SendDownloadFailed(context.Background(), order.ICCID, err.Error())
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to order profile", "message": err.Error()})
			return
		}

		profile := &repository.Profile{
			ICCID:       order.ICCID,
			EID:         order.EID,
			IMSI:        order.IMSI,
			MCC:         order.MCC,
			MNC:         order.MNC,
			ProfileType: order.ProfileType,
			State:       "provisioned",
			TenantID:    c.GetHeader("X-Tenant-ID"),
		}
		if err := repo.Create(c.Request.Context(), profile); err != nil {
			Logger.Warn().Err(err).Str("iccid", order.ICCID).Msg("Profile repo write failed")
		}

		// Send webhook notification for successful download
		if webhookClient != nil {
			go webhookClient.SendProfileDownloaded(context.Background(), order.ICCID, map[string]any{
				"imsi":        order.IMSI,
				"profileType": order.ProfileType,
				"status":      downloadResp.ExecutionStatus,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"profile": &ProfileResponse{
				ExecutionStatus: downloadResp.ExecutionStatus,
				StatusMessage:   downloadResp.StatusMessage,
				ProfileID:       order.ICCID,
			},
		})
	}
}

// GetProfileHandlerWithRepo returns the stored profile enriched with live ES2+ status.
func GetProfileHandlerWithRepo(client *es2.ES2Client, repo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		profileID := c.Param("profileId")
		if profileID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing profile ID"})
			return
		}

		stored, err := repo.Get(c.Request.Context(), profileID)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			Logger.Error().Err(err).Str("profile_id", profileID).Msg("Failed to read profile from store")
		}

		statusResp, statusErr := client.GetProfileStatus(context.Background(), &es2.GetProfileStatusRequest{
			ICCID: profileID,
			EID:   safeEID(stored),
		})

		payload := gin.H{
			"profileId": profileID,
			"checkedAt": time.Now().UTC(),
		}
		if stored != nil {
			payload["state"] = stored.State
			payload["imsi"] = stored.IMSI
			payload["tenantId"] = stored.TenantID
			payload["createdAt"] = stored.CreatedAt
			payload["updatedAt"] = stored.UpdatedAt
		}
		if statusErr != nil {
			Logger.Warn().Err(statusErr).Str("profile_id", profileID).Msg("Live profile status unavailable")
			payload["liveStatusError"] = statusErr.Error()
		} else {
			payload["executionStatus"] = statusResp.ExecutionStatus
			payload["profileState"] = statusResp.ProfileState
			payload["statusMessage"] = statusResp.StatusMessage
		}

		if stored == nil && statusErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found", "profileId": profileID})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "profile": payload})
	}
}

// ListProfilesHandlerWithRepo returns a paginated list of stored profiles.
func ListProfilesHandlerWithRepo(repo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := parsePositiveInt(c.Query("page"), 1)
		limit := min(parsePositiveInt(c.Query("limit"), 20), 100)

		Logger.Info().Int("page", page).Int("limit", limit).Msg("Listing eSIM profiles")

		profiles, total, err := repo.List(c.Request.Context(), repository.ListFilter{
			TenantID: c.Query("tenantId"),
			State:    c.Query("state"),
			Limit:    limit,
			Offset:   (page - 1) * limit,
		})
		if err != nil {
			Logger.Error().Err(err).Msg("Failed to list profiles")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list profiles", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"profiles": profiles,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})
	}
}

// DeleteProfileHandlerWithRepo deletes a profile from the SM-DP+ and marks it deleted in the repo.
func DeleteProfileHandlerWithRepo(client *es2.ES2Client, repo repository.ProfileRepository, webhookClient *webhook.WebhookClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		profileID := c.Param("profileId")
		if profileID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing profile ID"})
			return
		}

		deleteResp, err := client.DeleteProfile(context.Background(), &es2.DeleteProfileRequest{ICCID: profileID})
		if err != nil {
			Logger.Error().Err(err).Str("profile_id", profileID).Msg("Failed to delete profile upstream")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile", "message": err.Error()})
			return
		}

		if _, err := repo.UpdateState(c.Request.Context(), profileID, "deleted"); err != nil && !errors.Is(err, repository.ErrNotFound) {
			Logger.Warn().Err(err).Str("profile_id", profileID).Msg("Profile repo update failed")
		}

		// Send webhook notification for profile deletion
		if webhookClient != nil {
			go webhookClient.SendProfileDeleted(context.Background(), profileID, map[string]any{
				"executionStatus": deleteResp.ExecutionStatus,
				"statusMessage":   deleteResp.StatusMessage,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success":         true,
			"message":         "Profile deletion requested",
			"executionStatus": deleteResp.ExecutionStatus,
			"statusMessage":   deleteResp.StatusMessage,
		})
	}
}

// safeEID returns the EID from a stored profile or empty string if nil.
func safeEID(p *repository.Profile) string {
	if p == nil {
		return ""
	}
	return p.EID
}

// unused import guard (strconv used indirectly by parsePositiveInt in handlers.go)
var _ = strconv.Atoi
