"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Activity, AlertTriangle, TrendingUp, TrendingDown, Eye, RefreshCw } from "lucide-react";
import { api, Alert, MetricSample } from "@/lib/api";
import { ErrorAlert } from "@/components/ui/error-alert";
import { LoadingState, ActionLoading } from "@/components/ui/loading-state";

export default function MonitoringPage() {
  const [metrics, setMetrics] = useState<MetricSample[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [selectedQuery, setSelectedQuery] = useState("up");
  const [customQuery, setCustomQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const predefinedQueries = [
    { label: "System Uptime", query: "up" },
    { label: "CPU Usage", query: "rate(cpu_usage_total[5m])" },
    { label: "Memory Usage", query: "rate(memory_usage_bytes[5m])" },
    { label: "Network Traffic", query: "rate(network_traffic_bytes[5m])" },
    { label: "Active Connections", query: "active_connections_total" },
    { label: "Response Time", query: "http_request_duration_seconds" },
    { label: "Error Rate", query: "rate(http_requests_total{status=~\"5..\"}[5m])" },
    { label: "Request Rate", query: "rate(http_requests_total[5m])" }
  ];

  useEffect(() => {
    loadMonitoringData();
  }, [selectedQuery]);

  const loadMonitoringData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [metricsData, alertsData] = await Promise.all([
        api.monitoring.metrics(selectedQuery),
        api.monitoring.alerts()
      ]);
      
      setMetrics(metricsData.data);
      setAlerts(alertsData);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to load monitoring data";
      setError(errorMessage);
      console.error("Failed to load monitoring data:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleCustomQuery = async () => {
    if (!customQuery.trim()) return;
    
    try {
      setLoading(true);
      setError(null);
      const metricsData = await api.monitoring.metrics(customQuery);
      setMetrics(metricsData.data);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to execute custom query";
      setError(errorMessage);
      console.error("Failed to execute custom query:", err);
    } finally {
      setLoading(false);
    }
  };

  const getAlertSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical': return 'destructive';
      case 'warning': return 'warning';
      case 'info': return 'secondary';
      default: return 'secondary';
    }
  };

  const getAlertStatusColor = (state: string) => {
    switch (state.toLowerCase()) {
      case 'firing': return 'destructive';
      case 'resolved': return 'success';
      default: return 'secondary';
    }
  };

  const formatMetricValue = (value: number) => {
    if (value >= 1000000) return `${(value / 1000000).toFixed(2)}M`;
    if (value >= 1000) return `${(value / 1000).toFixed(2)}K`;
    return value.toFixed(2);
  };

  if (loading && metrics.length === 0) {
    return (
      <div className="p-8 space-y-6">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-48 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-96"></div>
        </div>
        <div className="grid gap-6 lg:grid-cols-2">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="animate-pulse">
              <div className="h-64 bg-gray-200 rounded"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h3 className="text-red-800 font-medium">Error loading monitoring data</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Activity className="size-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Monitoring</h1>
            <p className="text-muted-foreground mt-1">Advanced metrics and alerting dashboard.</p>
          </div>
        </div>
        <Button onClick={loadMonitoringData} disabled={loading}>
          <RefreshCw className={`size-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* Query Controls */}
      <Card>
        <CardHeader>
          <CardTitle>Metric Query</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex flex-wrap gap-2">
              {predefinedQueries.map((pq) => (
                <Button
                  key={pq.query}
                  variant={selectedQuery === pq.query ? "default" : "outline"}
                  size="sm"
                  onClick={() => setSelectedQuery(pq.query)}
                >
                  {pq.label}
                </Button>
              ))}
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                placeholder="Enter custom PromQL query..."
                value={customQuery}
                onChange={(e) => setCustomQuery(e.target.value)}
                className="flex-1 h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
              />
              <Button onClick={handleCustomQuery} disabled={!customQuery.trim() || loading}>
                Execute
              </Button>
            </div>
            <div className="text-sm text-muted-foreground">
              Current query: <code className="bg-muted px-2 py-1 rounded">{selectedQuery}</code>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Metrics Display */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="size-5" />
              Metric Values
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {metrics.length > 0 ? (
                metrics.map((metric, index) => (
                  <div key={index} className="flex items-center justify-between p-3 border rounded">
                    <div>
                      <p className="font-medium text-sm">
                        {new Date(metric.timestamp).toLocaleString()}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {metric.timestamp}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="font-mono text-lg">
                        {formatMetricValue(metric.value)}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {metric.value > 0 ? (
                          <TrendingUp className="inline size-3 text-green-500" />
                        ) : (
                          <TrendingDown className="inline size-3 text-red-500" />
                        )}
                      </p>
                    </div>
                  </div>
                ))
              ) : (
                <p className="text-muted-foreground text-center py-8">No metrics data available</p>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertTriangle className="size-5" />
              Active Alerts
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {alerts.length > 0 ? (
                alerts.map((alert, index) => (
                  <div key={index} className="border rounded p-3">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium text-sm">{alert.name}</span>
                      <div className="flex gap-2">
                        <Badge variant={getAlertSeverityColor(alert.severity)}>
                          {alert.severity}
                        </Badge>
                        <Badge variant={getAlertStatusColor(alert.state)}>
                          {alert.state}
                        </Badge>
                      </div>
                    </div>
                    <p className="text-sm text-muted-foreground mb-1">{alert.summary}</p>
                    <p className="text-xs text-muted-foreground">{alert.description}</p>
                    <div className="flex items-center justify-between mt-2 text-xs text-muted-foreground">
                      <span>Started: {new Date(alert.startsAt).toLocaleString()}</span>
                      {alert.endsAt && (
                        <span>Ended: {new Date(alert.endsAt).toLocaleString()}</span>
                      )}
                    </div>
                  </div>
                ))
              ) : (
                <p className="text-muted-foreground text-center py-8">No active alerts</p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* System Health Overview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Eye className="size-5" />
            System Health Overview
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <div className="text-center p-4 border rounded">
              <div className="text-2xl font-bold text-green-600">
                {alerts.filter(a => a.state === 'resolved').length}
              </div>
              <p className="text-sm text-muted-foreground">Resolved Alerts</p>
            </div>
            <div className="text-center p-4 border rounded">
              <div className="text-2xl font-bold text-red-600">
                {alerts.filter(a => a.state === 'firing').length}
              </div>
              <p className="text-sm text-muted-foreground">Active Alerts</p>
            </div>
            <div className="text-center p-4 border rounded">
              <div className="text-2xl font-bold text-orange-600">
                {alerts.filter(a => a.severity === 'critical').length}
              </div>
              <p className="text-sm text-muted-foreground">Critical Alerts</p>
            </div>
            <div className="text-center p-4 border rounded">
              <div className="text-2xl font-bold text-blue-600">
                {metrics.length}
              </div>
              <p className="text-sm text-muted-foreground">Data Points</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
