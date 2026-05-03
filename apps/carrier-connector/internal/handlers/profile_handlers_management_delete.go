package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
)

// DeleteProfileHandler deletes a profile
func DeleteProfileHandler(repo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		iccid := c.Param("iccid")
		if iccid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ICCID is required"})
			return
		}

		// Check if profile exists
		_, err := repo.Get(context.Background(), iccid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Profile not found",
				"message": err.Error(),
			})
			return
		}

		if err := repo.Delete(context.Background(), iccid); err != nil {
			Logger.Error().Err(err).Str("iccid", iccid).Msg("Failed to delete profile")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete profile",
				"message": err.Error(),
			})
			return
		}

		Logger.Info().
			Str("iccid", iccid).
			Msg("Profile deleted successfully")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Profile deleted successfully",
		})
	}
}

// GetProfileStatsHandler returns profile statistics
func GetProfileStatsHandler(repo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get all profiles to calculate stats
		filter := repository.ListFilter{
			Limit: 1000, // Get up to 1000 profiles for stats
		}

		profiles, total, err := repo.List(context.Background(), filter)
		if err != nil {
			Logger.Error().Err(err).Msg("Failed to get profile statistics")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get statistics",
				"message": err.Error(),
			})
			return
		}

		// Calculate basic stats
		stats := map[string]any{
			"total_profiles": total,
		}

		// Count by state
		stateCounts := make(map[string]int)
		for _, profile := range profiles {
			stateCounts[profile.State]++
		}
		stats["by_state"] = stateCounts

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"stats":   stats,
		})
	}
}
