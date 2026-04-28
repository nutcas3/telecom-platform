package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// PrometheusService queries a Prometheus HTTP API for real system metrics
// and alerts. Configure via PROMETHEUS_URL (e.g. http://prometheus:9090).
type PrometheusService struct {
	baseURL string
	http    *http.Client
}

func NewPrometheusService() *PrometheusService {
	base := os.Getenv("PROMETHEUS_URL")
	if base == "" {
		base = "http://prometheus:9090"
	}
	return &PrometheusService{
		baseURL: strings.TrimRight(base, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Available reports whether the configured Prometheus instance is reachable.
func (p *PrometheusService) Available(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/-/ready", nil)
	if err != nil {
		return false
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

type promResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error,omitempty"`
}

type PromVectorSample struct {
	Metric map[string]string `json:"metric"`
	Value  [2]any            `json:"value"`
}

type promVectorData struct {
	ResultType string             `json:"resultType"`
	Result     []PromVectorSample `json:"result"`
}

// Query runs an instant PromQL query.
func (p *PrometheusService) Query(ctx context.Context, promQL string) ([]PromVectorSample, error) {
	u := fmt.Sprintf("%s/api/v1/query?query=%s", p.baseURL, url.QueryEscape(promQL))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("prometheus query failed: %d %s", resp.StatusCode, string(b))
	}
	var out promResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, fmt.Errorf("prometheus query error: %s", out.Error)
	}
	var data promVectorData
	if err := json.Unmarshal(out.Data, &data); err != nil {
		return nil, err
	}
	return data.Result, nil
}

// Alert represents an active Prometheus alert.
type PromAlert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    time.Time         `json:"activeAt"`
	Value       string            `json:"value"`
}

type promAlertsData struct {
	Alerts []PromAlert `json:"alerts"`
}

// Alerts returns currently firing/pending alerts.
func (p *PrometheusService) Alerts(ctx context.Context) ([]PromAlert, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/api/v1/alerts", nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("prometheus alerts failed: %d %s", resp.StatusCode, string(b))
	}
	var out promResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, fmt.Errorf("prometheus alerts error: %s", out.Error)
	}
	var data promAlertsData
	if err := json.Unmarshal(out.Data, &data); err != nil {
		return nil, err
	}
	return data.Alerts, nil
}
