package handlers

import (
	"context"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
)

// UpdateProfileHandler updates a profile status
func UpdateProfileHandler(repo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		iccid := c.Param("iccid")
		if iccid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ICCID is required"})
			return
		}

		var update struct {
			State string `json:"status"`
		}

		if err := c.ShouldBindJSON(&update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Validate status (using State field)
		validStates := []string{"downloaded", "activated", "deactivated", "expired"}
		isValid := slices.Contains(validStates, update.State)

		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid status",
				"message": "Status must be one of: downloaded, activated, deactivated, expired",
			})
			return
		}

		// Use UpdateState method
		updatedProfile, err := repo.UpdateState(context.Background(), iccid, update.State)
		if err != nil {
			Logger.Error().Err(err).Str("iccid", iccid).Msg("Failed to update profile")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update profile",
				"message": err.Error(),
			})
			return
		}

		Logger.Info().
			Str("iccid", iccid).
			Str("status", update.State).
			Msg("Profile status updated successfully")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Profile updated successfully",
			"profile": updatedProfile,
		})
	}
}
