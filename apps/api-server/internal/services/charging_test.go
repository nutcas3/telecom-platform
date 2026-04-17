package services

import (
	"context"
	"testing"
)

func TestGetSystemStats(t *testing.T) {
	cs := &ChargingService{db: nil}
	stats, err := cs.GetSystemStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.ActiveSessions != 10 {
		t.Errorf("expected 10 active sessions, got %d", stats.ActiveSessions)
	}
	if stats.TotalAccounts != 100 {
		t.Errorf("expected 100 total accounts, got %d", stats.TotalAccounts)
	}
}

func TestGetHealthStatus(t *testing.T) {
	cs := &ChargingService{db: nil}
	health, err := cs.GetHealthStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health == nil {
		t.Fatal("expected non-nil health")
	}
	if !health.RedisConnected {
		t.Error("expected redis_connected=true")
	}
}

func TestGetUsageStats(t *testing.T) {
	cs := &ChargingService{db: nil}
	stats, err := cs.GetUsageStats(context.Background(), "208930000000001", "DAILY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Period != "DAILY" {
		t.Errorf("expected period=DAILY, got %s", stats.Period)
	}
	if stats.Trend == nil {
		t.Fatal("expected non-nil trend")
	}
}

func TestGetRealTimeUsage(t *testing.T) {
	cs := &ChargingService{db: nil}
	usage, err := cs.GetRealTimeUsage(context.Background(), "208930000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage.CurrentSession == nil {
		t.Fatal("expected non-nil current session")
	}
	if usage.TodayUsage == nil {
		t.Fatal("expected non-nil today usage")
	}
}

func TestListUsageEvents(t *testing.T) {
	cs := &ChargingService{db: nil}
	events, total, err := cs.ListUsageEvents(context.Background(), 10, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Error("expected at least 1 event")
	}
	if len(events) < 1 {
		t.Error("expected at least 1 event in slice")
	}
}

func TestSearchUsageEvents(t *testing.T) {
	cs := &ChargingService{db: nil}
	events, err := cs.SearchUsageEvents(context.Background(), "data", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) < 1 {
		t.Error("expected at least 1 event")
	}
}

func TestTriggerMaintenance(t *testing.T) {
	cs := &ChargingService{db: nil}
	ok, err := cs.TriggerMaintenance(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected maintenance trigger to succeed")
	}
}
