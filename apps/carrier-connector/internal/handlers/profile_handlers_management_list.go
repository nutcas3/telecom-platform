package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
)

// GetProfileHandler retrieves a profile by ICCID
func GetProfileHandler(repo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		iccid := c.Param("iccid")
		if iccid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ICCID is required"})
			return
		}

		profile, err := repo.Get(context.Background(), iccid)
		if err != nil {
			Logger.Error().Err(err).Str("iccid", iccid).Msg("Failed to retrieve profile")
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Profile not found",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"profile": profile,
		})
	}
}

// ListProfilesHandler retrieves all profiles with pagination
func ListProfilesHandler(repo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 50
		}

		offset := (page - 1) * limit

		// Use List method with filter
		filter := repository.ListFilter{
			Limit:  limit,
			Offset: offset,
		}

		profiles, total, err := repo.List(context.Background(), filter)
		if err != nil {
			Logger.Error().Err(err).Msg("Failed to retrieve profiles")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve profiles",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"profiles": profiles,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + limit - 1) / limit,
			},
		})
	}
}
