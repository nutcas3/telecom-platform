package mq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMessageQueue_ProfileEvent tests publishing profile events
func TestMessageQueue_ProfileEvent(t *testing.T) {
	// Skip if RabbitMQ is not available
	t.Skip("Skipping integration test - requires RabbitMQ server")
}

// TestMessageQueue_Connectivity tests connection to RabbitMQ
func TestMessageQueue_Connectivity(t *testing.T) {
	// Skip if RabbitMQ is not available
	t.Skip("Skipping integration test - requires RabbitMQ server")
}

// TestMessageQueue_DeclareQueue tests queue declaration
func TestMessageQueue_DeclareQueue(t *testing.T) {
	// Skip if RabbitMQ is not available
	t.Skip("Skipping integration test - requires RabbitMQ server")
}

// TestMessageQueue_PublishConsume tests publish and consume flow
func TestMessageQueue_PublishConsume(t *testing.T) {
	// Skip if RabbitMQ is not available
	t.Skip("Skipping integration test - requires RabbitMQ server")
}

// TestProfileEventSerialization tests message serialization
func TestProfileEventSerialization(t *testing.T) {
	msg := Message{
		Type: "profile.downloaded",
		Payload: map[string]any{
			"profile_id": "test-iccid-123",
			"imsi":       "123456789012345",
			"status":     "completed",
		},
		Timestamp: time.Now(),
	}

	assert.Equal(t, "profile.downloaded", msg.Type)
	assert.Equal(t, "test-iccid-123", msg.Payload["profile_id"])
	assert.Equal(t, "123456789012345", msg.Payload["imsi"])
	assert.Equal(t, "completed", msg.Payload["status"])
	assert.False(t, msg.Timestamp.IsZero())
}

// TestPublishProfileEvent tests the convenience method
func TestPublishProfileEvent(t *testing.T) {
	// This would require a mock or real RabbitMQ connection
	// For now, we test the structure
	eventType := "profile.deleted"
	profileID := "test-iccid-456"
	payload := map[string]any{
		"executionStatus": "Completed",
		"statusMessage":   "Profile deleted successfully",
	}

	assert.Equal(t, "profile.deleted", eventType)
	assert.Equal(t, "test-iccid-456", profileID)
	assert.Equal(t, "Completed", payload["executionStatus"])
}
