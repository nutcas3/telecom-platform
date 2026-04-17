package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *ChargingEngineClient) {
	srv := httptest.NewServer(handler)
	client := NewChargingEngineClient(&config.ChargingEngineConfig{
		BaseURL: srv.URL,
		Timeout: 5 * time.Second,
	})
	return srv, client
}

func TestCheckCredit_Allowed(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/credit/10.0.0.1/check" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req CreditCheckRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.BytesRequested != 1000 {
			t.Errorf("expected 1000 bytes, got %d", req.BytesRequested)
		}
		json.NewEncoder(w).Encode(CreditCheckResponse{Allowed: true, RemainingBytes: 5000})
	})
	defer srv.Close()

	resp, err := client.CheckCredit(context.Background(), "10.0.0.1", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Allowed {
		t.Error("expected allowed=true")
	}
	if resp.RemainingBytes != 5000 {
		t.Errorf("expected remaining=5000, got %d", resp.RemainingBytes)
	}
}

func TestCheckCredit_ServerError(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	_, err := client.CheckCredit(context.Background(), "10.0.0.1", 1000)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestDeductCredit(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/credit/10.0.0.2/deduct" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req DeductRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.BytesUsed != 500 {
			t.Errorf("expected 500, got %d", req.BytesUsed)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	err := client.DeductCredit(context.Background(), "10.0.0.2", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddCredit(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/credit/10.0.0.3/add" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req AddCreditRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.BytesToAdd != 2000 {
			t.Errorf("expected 2000, got %d", req.BytesToAdd)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	err := client.AddCredit(context.Background(), "10.0.0.3", 2000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetBalance(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(BalanceResponse{IP: "10.0.0.4", BalanceBytes: 9999})
	})
	defer srv.Close()

	resp, err := client.GetBalance(context.Background(), "10.0.0.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.BalanceBytes != 9999 {
		t.Errorf("expected 9999, got %d", resp.BalanceBytes)
	}
}

func TestHealthCheck(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(EngineHealthResponse{Status: "ok", Timestamp: "2026-04-14T12:00:00Z"})
	})
	defer srv.Close()

	resp, err := client.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status=ok, got %s", resp.Status)
	}
}

func TestContextCancellation(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GetBalance(ctx, "10.0.0.1")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
