"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertTriangle, Wrench, Clock, CheckCircle, XCircle, Server, Database, Wifi } from "lucide-react";
import { apiClient, MaintenanceMetrics } from "@/lib/api-client";

interface Asset {
  id: string;
  name: string;
  type: string;
  health: number;
  status: string;
  lastMaintenance: string;
  nextMaintenance?: string;
  predictedFailure?: string;
  riskFactors: string[];
}

export default function MaintenancePage() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [metrics, setMetrics] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Simulate API calls
    setTimeout(() => {
      setAssets([
        {
          id: "server-1",
          name: "Web Server 1",
          type: "server",
          health: 85,
          status: "healthy",
          lastMaintenance: "2026-04-15T10:30:00Z",
          nextMaintenance: "2026-06-15T10:30:00Z",
          riskFactors: [],
        },
        {
          id: "server-2",
          name: "Web Server 2",
          type: "server",
          health: 92,
          status: "healthy",
          lastMaintenance: "2026-04-20T14:15:00Z",
          nextMaintenance: "2026-06-20T14:15:00Z",
          riskFactors: [],
        },
        {
          id: "db-1",
          name: "Primary Database",
          type: "database",
          health: 78,
          status: "warning",
          lastMaintenance: "2026-03-10T09:00:00Z",
          nextMaintenance: "2026-05-10T09:00:00Z",
          riskFactors: ["High CPU usage", "Memory pressure"],
        },
        {
          id: "router-1",
          name: "Core Router",
          type: "network",
          health: 65,
          status: "warning",
          lastMaintenance: "2026-02-28T16:45:00Z",
          nextMaintenance: "2026-04-28T16:45:00Z",
          predictedFailure: "2026-09-01T00:00:00Z",
          riskFactors: ["Age", "Error rate increase", "Temperature spikes"],
        },
      ]);

      setMetrics({
        totalAssets: 1250,
        healthyAssets: 1180,
        assetsNeedingAttention: 70,
        uptime: 99.95,
        meanTimeToFailure: 8760,
        meanTimeToRepair: 4,
        byType: {
          server: 450,
          database: 125,
          network: 300,
          storage: 375,
        },
        byStatus: {
          healthy: 1180,
          warning: 60,
          critical: 10,
        },
      });
      setLoading(false);
    }, 1000);
  }, []);

  const getHealthColor = (health: number) => {
    if (health >= 90) return "text-green-600";
    if (health >= 75) return "text-yellow-600";
    if (health >= 60) return "text-orange-600";
    return "text-red-600";
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case "healthy": return "default";
      case "warning": return "secondary";
      case "critical": return "destructive";
      default: return "outline";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "healthy": return <CheckCircle className="h-4 w-4 text-green-500" />;
      case "warning": return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case "critical": return <AlertTriangle className="h-4 w-4 text-red-500" />;
      default: return <Clock className="h-4 w-4 text-gray-500" />;
    }
  };

  const getAssetIcon = (type: string) => {
    switch (type) {
      case "server": return <Server className="h-4 w-4" />;
      case "database": return <Database className="h-4 w-4" />;
      case "network": return <Wifi className="h-4 w-4" />;
      default: return <Wrench className="h-4 w-4" />;
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
          <Wrench className="mr-3 h-8 w-8 text-blue-600" />
          Predictive Maintenance
        </h1>
        <Button>
          <Wrench className="mr-2 h-4 w-4" />
          Schedule Maintenance
        </Button>
      </div>

      {/* Maintenance Metrics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Assets</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.totalAssets.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">Infrastructure assets</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Healthy Assets</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{metrics?.healthyAssets.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">Operating normally</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Uptime</CardTitle>
            <Wifi className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{metrics?.uptime}%</div>
            <p className="text-xs text-muted-foreground">Excellent availability</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">MTTR</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.meanTimeToRepair}h</div>
            <p className="text-xs text-muted-foreground">Mean time to repair</p>
          </CardContent>
        </Card>
      </div>

      {/* Asset Distribution */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Assets by Type</CardTitle>
            <CardDescription>Infrastructure asset distribution</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {Object.entries(metrics?.byType || {}).map(([type, count]) => (
              <div key={type} className="flex justify-between items-center">
                <div className="flex items-center space-x-2">
                  {getAssetIcon(type)}
                  <span className="capitalize">{type}</span>
                </div>
                <Badge variant="outline">{count?.toLocaleString()}</Badge>
              </div>
            ))}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Assets by Status</CardTitle>
            <CardDescription>Current health status distribution</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {Object.entries(metrics?.byStatus || {}).map(([status, count]) => (
              <div key={status} className="flex justify-between items-center">
                <div className="flex items-center space-x-2">
                  {getStatusIcon(status)}
                  <span className="capitalize">{status}</span>
                </div>
                <Badge variant={getStatusBadgeVariant(status) as any}>{count?.toLocaleString()}</Badge>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      {/* Assets Health Table */}
      <Card>
        <CardHeader>
          <CardTitle>Asset Health Status</CardTitle>
          <CardDescription>Real-time health monitoring of critical infrastructure assets</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Asset</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Health</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Last Maintenance</TableHead>
                <TableHead>Next Maintenance</TableHead>
                <TableHead>Predicted Failure</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {assets.map((asset) => (
                <TableRow key={asset.id}>
                  <TableCell className="font-medium">{asset.name}</TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                      {getAssetIcon(asset.type)}
                      <span className="capitalize">{asset.type}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                      <span className={`font-semibold ${getHealthColor(asset.health)}`}>
                        {asset.health}%
                      </span>
                      <Progress value={asset.health} className="h-2 w-16" />
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(asset.status)}
                      <Badge variant={getStatusBadgeVariant(asset.status) as any}>
                        {asset.status}
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell>
                    {new Date(asset.lastMaintenance).toLocaleDateString()}
                  </TableCell>
                  <TableCell>
                    {asset.nextMaintenance ? new Date(asset.nextMaintenance).toLocaleDateString() : "N/A"}
                  </TableCell>
                  <TableCell>
                    {asset.predictedFailure ? (
                      <Badge variant="destructive">
                        {new Date(asset.predictedFailure).toLocaleDateString()}
                      </Badge>
                    ) : (
                      <span className="text-muted-foreground">N/A</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <Button size="sm" variant="outline">
                      Maintain
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Risk Factors and Predictions */}
      <div className="grid gap-4 md:grid-cols-2">
        {assets.filter(a => a.riskFactors.length > 0).map((asset) => (
          <Card key={asset.id}>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                {asset.name}
                <Badge variant={getStatusBadgeVariant(asset.status) as any}>{asset.status}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <h4 className="font-semibold mb-2">Risk Factors</h4>
                <ul className="text-sm text-muted-foreground space-y-1">
                  {asset.riskFactors.map((factor, idx) => (
                    <li key={idx} className="flex items-start">
                      <span className="mr-2">•</span>
                      {factor}
                    </li>
                  ))}
                </ul>
              </div>
              {asset.predictedFailure && (
                <div>
                  <h4 className="font-semibold mb-2">Failure Prediction</h4>
                  <p className="text-sm text-muted-foreground">
                    Predicted failure on {new Date(asset.predictedFailure).toLocaleDateString()} with 82.5% confidence
                  </p>
                </div>
              )}
              <Button className="w-full" variant="outline">
                Schedule Maintenance
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Maintenance Recommendations */}
      <div className="grid gap-4 md:grid-cols-2">
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            <strong>Critical Maintenance Required:</strong> Core Router shows 65% health with predicted failure in September. Schedule immediate maintenance to prevent service disruption.
          </AlertDescription>
        </Alert>
        <Alert>
          <CheckCircle className="h-4 w-4" />
          <AlertDescription>
            <strong>Preventive Maintenance:</strong> Primary Database showing warning signs. Schedule maintenance within 30 days to maintain optimal performance.
          </AlertDescription>
        </Alert>
      </div>
    </div>
  );
}
