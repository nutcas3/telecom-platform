"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { TrendingDown, AlertTriangle, Users, Target } from "lucide-react";
import { apiClient, ChurnMetrics, ChurnPrediction } from "@/lib/api-client";

export default function ChurnPage() {
  const [predictions, setPredictions] = useState<ChurnPrediction[]>([]);
  const [metrics, setMetrics] = useState<ChurnMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedRiskLevel, setSelectedRiskLevel] = useState<string>("all");

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch churn metrics and predictions in parallel
        const [metricsResponse, predictionsResponse] = await Promise.all([
          apiClient.getChurnMetrics(),
          apiClient.getChurnPredictions(selectedRiskLevel === "all" ? undefined : selectedRiskLevel, 100),
        ]);

        if (metricsResponse.error || predictionsResponse.error) {
          setError('Failed to fetch some churn data');
        }

        setMetrics(metricsResponse.data || null);
        setPredictions(predictionsResponse.data || []);
      } catch (err) {
        console.error('Failed to fetch churn data:', err);
        setError('Failed to load churn analysis data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [selectedRiskLevel]);

  const filteredPredictions = predictions.filter(p => 
    selectedRiskLevel === "all" || p.riskLevel === selectedRiskLevel
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
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Error Loading Churn Analysis</h3>
          <p className="text-gray-600">{error}</p>
          <Button onClick={() => window.location.reload()} className="mt-4">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const getRiskBadgeVariant = (level: string) => {
    switch (level) {
      case "critical": return "destructive";
      case "high": return "destructive";
      case "medium": return "secondary";
      case "low": return "outline";
      default: return "secondary";
    }
  };

  const getRiskColor = (level: string) => {
    switch (level) {
      case "critical": return "text-red-600";
      case "high": return "text-orange-600";
      case "medium": return "text-yellow-600";
      case "low": return "text-green-600";
      default: return "text-gray-600";
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
          <TrendingDown className="mr-3 h-8 w-8 text-red-600" />
          Churn Analysis
        </h1>
        <Button>
          <Target className="mr-2 h-4 w-4" />
          Run Prediction Model
        </Button>
      </div>

      {/* Metrics Overview */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Subscribers</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.totalSubscribers.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">Active customers</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Churn Rate</CardTitle>
            <TrendingDown className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{metrics?.churnRate}%</div>
            <p className="text-xs text-muted-foreground">Monthly churn</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">At Risk Customers</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-orange-600">7,500</div>
            <p className="text-xs text-muted-foreground">High & critical risk</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg Tenure</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.averageTenureDays} days</div>
            <p className="text-xs text-muted-foreground">Customer lifetime</p>
          </CardContent>
        </Card>
      </div>

      {/* Risk Distribution */}
      <Card>
        <CardHeader>
          <CardTitle>Risk Distribution</CardTitle>
          <CardDescription>Customer churn risk levels across the subscriber base</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {Object.entries(metrics?.riskDistribution || {}).map(([level, count]) => (
            <div key={level} className="space-y-2">
              <div className="flex justify-between items-center">
                <span className="capitalize font-medium">{level}</span>
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-muted-foreground">{count?.toLocaleString()}</span>
                  <Badge variant={getRiskBadgeVariant(level) as any}>{level}</Badge>
                </div>
              </div>
              <Progress 
                value={count && metrics?.totalSubscribers ? (count / metrics.totalSubscribers) * 100 : 0} 
                className="h-3"
              />
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Filter Controls */}
      <div className="flex items-center space-x-4">
        <span className="text-sm font-medium">Filter by risk level:</span>
        <div className="flex space-x-2">
          {["all", "critical", "high", "medium", "low"].map((level) => (
            <Button
              key={level}
              variant={selectedRiskLevel === level ? "default" : "outline"}
              size="sm"
              onClick={() => setSelectedRiskLevel(level)}
            >
              {level.charAt(0).toUpperCase() + level.slice(1)}
            </Button>
          ))}
        </div>
      </div>

      {/* At-Risk Customers Table */}
      <Card>
        <CardHeader>
          <CardTitle>At-Risk Customers</CardTitle>
          <CardDescription>Customers with elevated churn risk requiring attention</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Profile ID</TableHead>
                <TableHead>Risk Level</TableHead>
                <TableHead>Risk Score</TableHead>
                <TableHead>Predicted Churn</TableHead>
                <TableHead>Key Reasons</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredPredictions.map((prediction) => (
                <TableRow key={prediction.profileId}>
                  <TableCell className="font-medium">{prediction.profileId}</TableCell>
                  <TableCell>
                    <Badge variant={getRiskBadgeVariant(prediction.riskLevel) as any}>
                      {prediction.riskLevel}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <span className={`font-semibold ${getRiskColor(prediction.riskLevel)}`}>
                      {prediction.riskScore}%
                    </span>
                  </TableCell>
                  <TableCell>{prediction.predictedChurnDate || "N/A"}</TableCell>
                  <TableCell>
                    <div className="max-w-xs">
                      <p className="text-sm text-muted-foreground truncate">
                        {prediction.reasons.join(", ")}
                      </p>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Button size="sm" variant="outline">
                      Take Action
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Customer Retention Recommendations */}
      <div className="grid gap-4 md:grid-cols-2">
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            <strong>Critical Risk:</strong> 1,500 customers at immediate risk of churn. Recommend immediate outreach with personalized retention offers.
          </AlertDescription>
        </Alert>
        <Alert>
          <Target className="h-4 w-4" />
          <AlertDescription>
            <strong>Proactive Strategy:</strong> Focus on high-risk segment with targeted campaigns to reduce churn by 25% this quarter.
          </AlertDescription>
        </Alert>
      </div>
    </div>
  );
}
