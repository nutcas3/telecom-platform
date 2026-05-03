"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { AlertTriangle, Shield, Eye, CheckCircle, XCircle, Clock } from "lucide-react";
import { apiClient, FraudAlert, FraudMetrics } from "@/lib/api-client";

export default function FraudPage() {
  const [alerts, setAlerts] = useState<FraudAlert[]>([]);
  const [metrics, setMetrics] = useState<FraudMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedSeverity, setSelectedSeverity] = useState<string>("all");

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch fraud metrics and alerts in parallel
        const [metricsResponse, alertsResponse] = await Promise.all([
          apiClient.getFraudMetrics(),
          apiClient.getFraudAlerts(selectedSeverity === "all" ? undefined : selectedSeverity),
        ]);

        if (metricsResponse.error || alertsResponse.error) {
          setError('Failed to fetch some fraud data');
        }

        setMetrics(metricsResponse.data || null);
        setAlerts(alertsResponse.data || []);
      } catch (err) {
        console.error('Failed to fetch fraud data:', err);
        setError('Failed to load fraud detection data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [selectedSeverity]);

  const filteredAlerts = alerts.filter(a => 
    selectedSeverity === "all" || a.severity === selectedSeverity
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <AlertTriangle className="mx-auto h-12 w-12 text-red-500 mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Error Loading Fraud Detection</h3>
          <p className="text-gray-600">{error}</p>
          <Button onClick={() => window.location.reload()} className="mt-4">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const getSeverityBadgeVariant = (severity: string) => {
    switch (severity) {
      case "critical": return "destructive";
      case "high": return "destructive";
      case "medium": return "secondary";
      case "low": return "outline";
      default: return "secondary";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "new": return <Clock className="h-4 w-4 text-gray-500" />;
      case "investigating": return <Eye className="h-4 w-4 text-blue-500" />;
      case "resolved": return <CheckCircle className="h-4 w-4 text-green-500" />;
      case "blocked": return <XCircle className="h-4 w-4 text-red-500" />;
      default: return <Clock className="h-4 w-4 text-gray-500" />;
    }
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case "new": return "outline";
      case "investigating": return "secondary";
      case "resolved": return "default";
      case "blocked": return "destructive";
      default: return "outline";
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold flex items-center">
          <Shield className="mr-3 h-8 w-8 text-red-600" />
          Fraud Detection
        </h1>
        <Button>
          <AlertTriangle className="mr-2 h-4 w-4" />
          Run Security Scan
        </Button>
      </div>

      {/* Metrics Overview */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Alerts</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.totalAlerts.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">Last 30 days</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Resolution Rate</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{metrics?.resolutionRate}%</div>
            <p className="text-xs text-muted-foreground">Alerts resolved</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">False Positives</CardTitle>
            <XCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-orange-600">{metrics?.falsePositiveRate}%</div>
            <p className="text-xs text-muted-foreground">Accuracy improving</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Critical Alerts</CardTitle>
            <AlertTriangle className="h-4 w-4 text-red-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{metrics?.bySeverity.critical}</div>
            <p className="text-xs text-muted-foreground">Immediate action needed</p>
          </CardContent>
        </Card>
      </div>

      {/* Alert Types Distribution */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Alerts by Type</CardTitle>
            <CardDescription>Distribution of fraud detection alerts</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {Object.entries(metrics?.byType || {}).map(([type, count]) => (
              <div key={type} className="flex justify-between items-center">
                <span className="capitalize">{type.replace("_", " ")}</span>
                <Badge variant="outline">{count?.toLocaleString()}</Badge>
              </div>
            ))}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Alerts by Severity</CardTitle>
            <CardDescription>Risk level distribution</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {Object.entries(metrics?.bySeverity || {}).map(([severity, count]) => (
              <div key={severity} className="flex justify-between items-center">
                <span className="capitalize">{severity}</span>
                <Badge variant={getSeverityBadgeVariant(severity) as any}>{severity}</Badge>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      {/* Filter Controls */}
      <div className="flex items-center space-x-4">
        <span className="text-sm font-medium">Filter by severity:</span>
        <div className="flex space-x-2">
          {["all", "critical", "high", "medium", "low"].map((severity) => (
            <Button
              key={severity}
              variant={selectedSeverity === severity ? "default" : "outline"}
              size="sm"
              onClick={() => setSelectedSeverity(severity)}
            >
              {severity.charAt(0).toUpperCase() + severity.slice(1)}
            </Button>
          ))}
        </div>
      </div>

      {/* Recent Fraud Alerts */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Fraud Alerts</CardTitle>
          <CardDescription>Latest security alerts requiring attention</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Alert ID</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Severity</TableHead>
                <TableHead>Profile</TableHead>
                <TableHead>Risk Score</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredAlerts.map((alert) => (
                <TableRow key={alert.id}>
                  <TableCell className="font-medium">{alert.id}</TableCell>
                  <TableCell>
                    <Badge variant="outline">
                      {alert.type.replace("_", " ")}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={getSeverityBadgeVariant(alert.severity) as any}>
                      {alert.severity}
                    </Badge>
                  </TableCell>
                  <TableCell>{alert.profileId}</TableCell>
                  <TableCell>
                    <span className={`font-semibold ${
                      alert.riskScore >= 90 ? 'text-red-600' : 
                      alert.riskScore >= 75 ? 'text-orange-600' : 
                      alert.riskScore >= 50 ? 'text-yellow-600' : 'text-green-600'
                    }`}>
                      {alert.riskScore}%
                    </span>
                  </TableCell>
                  <TableCell>
                    <div className="max-w-xs">
                      <p className="text-sm truncate">{alert.description}</p>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(alert.status)}
                      <Badge variant={getStatusBadgeVariant(alert.status) as any}>
                        {alert.status}
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Button size="sm" variant="outline">
                      Review
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Security Recommendations */}
      <div className="grid gap-4 md:grid-cols-2">
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            <strong>Critical Alert:</strong> Account takeover attempt detected. Immediate action required to secure the account and prevent unauthorized access.
          </AlertDescription>
        </Alert>
        <Alert>
          <Shield className="h-4 w-4" />
          <AlertDescription>
            <strong>Security Tip:</strong> Consider implementing multi-factor authentication for high-risk accounts to prevent account takeover attempts.
          </AlertDescription>
        </Alert>
      </div>
    </div>
  );
}
