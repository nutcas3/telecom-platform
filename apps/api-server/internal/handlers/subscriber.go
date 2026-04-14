package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// SubscriberHandler handles subscriber-related HTTP requests
type SubscriberHandler struct {
	subscriberService *services.SubscriberService
}

// NewSubscriberHandler creates a new subscriber handler
func NewSubscriberHandler(subscriberService *services.SubscriberService) *SubscriberHandler {
	return &SubscriberHandler{
		subscriberService: subscriberService,
	}
}

// CreateSubscriber creates a new subscriber
// @Summary Create a new subscriber
// @Description Create a new subscriber with allocated IMSI
// @Tags subscribers
// @Accept json
// @Produce json
// @Param subscriber body services.CreateSubscriberRequest true "Subscriber data"
// @Success 201 {object} models.Subscriber
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribers [post]
func (h *SubscriberHandler) CreateSubscriber(c *gin.Context) {
	var req services.CreateSubscriberRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	subscriber, err := h.subscriberService.CreateSubscriber(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create subscriber",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, subscriber)
}

// GetSubscriber retrieves a subscriber by ID
// @Summary Get subscriber by ID
// @Description Retrieve subscriber information by ID
// @Tags subscribers
// @Produce json
// @Param id path int true "Subscriber ID"
// @Success 200 {object} models.Subscriber
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribers/{id} [get]
func (h *SubscriberHandler) GetSubscriber(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid subscriber ID",
			Details: err.Error(),
		})
		return
	}

	subscriber, err := h.subscriberService.GetSubscriber(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Subscriber not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, subscriber)
}

// GetSubscriberByIMSI retrieves a subscriber by IMSI
// @Summary Get subscriber by IMSI
// @Description Retrieve subscriber information by IMSI
// @Tags subscribers
// @Produce json
// @Param imsi path string true "IMSI"
// @Success 200 {object} models.Subscriber
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribers/imsi/{imsi} [get]
func (h *SubscriberHandler) GetSubscriberByIMSI(c *gin.Context) {
	imsi := models.IMSI(c.Param("imsi"))

	subscriber, err := h.subscriberService.GetSubscriberByIMSI(c.Request.Context(), imsi)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Subscriber not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, subscriber)
}

// UpdateSubscriber updates subscriber information
// @Summary Update subscriber
// @Description Update subscriber information
// @Tags subscribers
// @Accept json
// @Produce json
// @Param id path int true "Subscriber ID"
// @Param subscriber body services.UpdateSubscriberRequest true "Updated subscriber data"
// @Success 200 {object} models.Subscriber
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribers/{id} [put]
func (h *SubscriberHandler) UpdateSubscriber(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid subscriber ID",
			Details: err.Error(),
		})
		return
	}

	var req services.UpdateSubscriberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	subscriber, err := h.subscriberService.UpdateSubscriber(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update subscriber",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, subscriber)
}

// ListSubscribers lists subscribers with pagination and filtering
// @Summary List subscribers
// @Description List subscribers with pagination and filtering
// @Tags subscribers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Subscriber status"
// @Param organization_id query string false "Organization ID"
// @Param search query string false "Search term"
// @Success 200 {object} services.ListSubscribersResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribers [get]
func (h *SubscriberHandler) ListSubscribers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if pageSize > 100 {
		pageSize = 100
	}

	req := &services.ListSubscribersRequest{
		Page:           page,
		PageSize:       pageSize,
		Status:         models.SubscriberStatus(c.Query("status")),
		OrganizationID: c.Query("organization_id"),
		Search:         c.Query("search"),
	}

	response, err := h.subscriberService.ListSubscribers(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list subscribers",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SuspendSubscriber suspends a subscriber
// @Summary Suspend subscriber
// @Description Suspend a subscriber and terminate their sessions
// @Tags subscribers
// @Produce json
// @Param id path int true "Subscriber ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribers/{id}/suspend [post]
func (h *SubscriberHandler) SuspendSubscriber(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid subscriber ID",
			Details: err.Error(),
		})
		return
	}

	if err := h.subscriberService.SuspendSubscriber(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to suspend subscriber",
			Details: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// TerminateSubscriber terminates a subscriber
// @Summary Terminate subscriber
// @Description Terminate a subscriber and deactivate their eSIM profile
// @Tags subscribers
// @Produce json
// @Param id path int true "Subscriber ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscribers/{id}/terminate [post]
func (h *SubscriberHandler) TerminateSubscriber(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid subscriber ID",
			Details: err.Error(),
		})
		return
	}

	if err := h.subscriberService.TerminateSubscriber(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to terminate subscriber",
			Details: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
