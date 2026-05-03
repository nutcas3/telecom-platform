"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { BarChart3, TrendingUp, TrendingDown, AlertCircle } from "lucide-react";
import { apiClient, ChurnMetrics, MarketMetrics, PricingMetrics, MaintenanceMetrics } from "@/lib/api-client";

export default function AnalyticsPage() {
  const [metrics, setMetrics] = useState<{
    churn?: ChurnMetrics;
    market?: MarketMetrics;
    maintenance?: MaintenanceMetrics;
    pricing?: PricingMetrics;
  } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch all metrics in parallel
        const [churnResponse, marketResponse, maintenanceResponse, pricingResponse] = await Promise.all([
          apiClient.getChurnMetrics(),
          apiClient.getMarketMetrics(),
          apiClient.getMaintenanceMetrics(),
          apiClient.getPricingMetrics(),
        ]);

        if (churnResponse.error || marketResponse.error || maintenanceResponse.error || pricingResponse.error) {
          setError('Failed to fetch some metrics');
        }

        setMetrics({
          churn: churnResponse.data,
          market: marketResponse.data,
          maintenance: maintenanceResponse.data,
          pricing: pricingResponse.data,
        });
      } catch (err) {
        console.error('Failed to fetch analytics metrics:', err);
        setError('Failed to load analytics data');
      } finally {
        setLoading(false);
      }
    };

    fetchMetrics();
  }, []);

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
          <AlertCircle className="mx-auto h-12 w-12 text-red-500 mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Error Loading Analytics</h3>
          <p className="text-gray-600">{error}</p>
          <Button onClick={() => window.location.reload()} className="mt-4">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Analytics Dashboard</h1>
        <Button>
          <BarChart3 className="mr-2 h-4 w-4" />
          Export Report
        </Button>
      </div>

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="churn">Churn Analysis</TabsTrigger>
          <TabsTrigger value="market">Market Analytics</TabsTrigger>
          <TabsTrigger value="maintenance">Maintenance</TabsTrigger>
          <TabsTrigger value="pricing">Pricing</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Subscribers</CardTitle>
                <BarChart3 className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{metrics?.churn?.totalSubscribers?.toLocaleString() || 'N/A'}</div>
                <p className="text-xs text-muted-foreground">+12.5% from last month</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Churn Rate</CardTitle>
                <TrendingDown className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{metrics?.churn?.churnRate ? `${metrics.churn.churnRate}%` : 'N/A'}</div>
                <p className="text-xs text-muted-foreground">-0.3% from last month</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Market Share</CardTitle>
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{metrics?.market?.marketShare ? `${(metrics.market.marketShare * 100).toFixed(3)}%` : 'N/A'}</div>
                <p className="text-xs text-muted-foreground">+0.1% from last month</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">System Uptime</CardTitle>
                <AlertCircle className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{metrics?.maintenance?.uptime ? `${metrics.maintenance.uptime}%` : 'N/A'}</div>
                <p className="text-xs text-muted-foreground">Excellent</p>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="churn" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Churn Metrics</CardTitle>
                <CardDescription>Customer churn analysis for the current period</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between">
                  <span>Total Subscribers</span>
                  <span className="font-semibold">{metrics?.churn?.totalSubscribers?.toLocaleString() || 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Churned Subscribers</span>
                  <span className="font-semibold text-red-600">{metrics?.churn?.churnedSubscribers?.toLocaleString() || 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Monthly Churn Rate</span>
                  <span className="font-semibold">{metrics?.churn?.monthlyChurnRate ? `${metrics.churn.monthlyChurnRate}%` : 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Average Tenure</span>
                  <span className="font-semibold">{metrics?.churn?.averageTenureDays ? `${metrics.churn.averageTenureDays} days` : 'N/A'}</span>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Risk Distribution</CardTitle>
                <CardDescription>Customers at risk of churn</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {Object.entries(metrics?.churn?.riskDistribution || {}).map(([level, count]) => (
                  <div key={level} className="space-y-2">
                    <div className="flex justify-between">
                      <span className="capitalize">{level}</span>
                      <span>{count?.toLocaleString() || '0'}</span>
                    </div>
                    <Progress 
                      value={count && metrics?.churn?.totalSubscribers ? (count / metrics.churn.totalSubscribers) * 100 : 0} 
                      className="h-2"
                    />
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="market" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Market Metrics</CardTitle>
                <CardDescription>Market penetration and growth</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between">
                  <span>Total Market Size</span>
                  <span className="font-semibold">{metrics?.market?.totalMarketSize ? `${(metrics.market.totalMarketSize / 1000000000).toFixed(1)}B` : 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Our Subscribers</span>
                  <span className="font-semibold">{metrics?.market?.ourSubscribers?.toLocaleString() || 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Market Share</span>
                  <span className="font-semibold">{metrics?.market?.marketShare ? `${(metrics.market.marketShare * 100).toFixed(3)}%` : 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Growth Rate</span>
                  <span className="font-semibold text-green-600">{metrics?.market?.growthRate ? `+${metrics.market.growthRate}%` : 'N/A'}</span>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Market by Country</CardTitle>
                <CardDescription>Performance across regions</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {Object.entries(metrics?.market?.byCountry || {}).map(([country, data]: [string, any]) => (
                  <div key={country} className="space-y-2">
                    <div className="flex justify-between">
                      <span>{country}</span>
                      <span>{data?.penetration ? `${(data.penetration * 100).toFixed(2)}%` : 'N/A'}</span>
                    </div>
                    <Progress 
                      value={data?.penetration ? data.penetration * 100 : 0} 
                      className="h-2"
                    />
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="maintenance" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>System Health</CardTitle>
                <CardDescription>Infrastructure maintenance metrics</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between">
                  <span>Total Assets</span>
                  <span className="font-semibold">{metrics?.maintenance?.totalAssets || 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Healthy Assets</span>
                  <span className="font-semibold text-green-600">{metrics?.maintenance?.healthyAssets || 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Assets Needing Attention</span>
                  <span className="font-semibold text-yellow-600">{metrics?.maintenance?.assetsNeedingAttention || 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>System Uptime</span>
                  <span className="font-semibold">{metrics?.maintenance?.uptime ? `${metrics.maintenance.uptime}%` : 'N/A'}</span>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Performance Metrics</CardTitle>
                <CardDescription>Mean time metrics</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between">
                  <span>Mean Time To Failure</span>
                  <span className="font-semibold">{metrics?.maintenance?.meanTimeToFailure ? `${(metrics.maintenance.meanTimeToFailure / 24 / 365).toFixed(1)} years` : 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Mean Time To Repair</span>
                  <span className="font-semibold">{metrics?.maintenance?.meanTimeToRepair ? `${metrics.maintenance.meanTimeToRepair} hours` : 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Asset Health Score</span>
                  <Badge variant="secondary">Excellent</Badge>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="pricing" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Revenue Metrics</CardTitle>
                <CardDescription>Pricing and revenue analysis</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between">
                  <span>Total Revenue</span>
                  <span className="font-semibold">{metrics?.pricing?.totalRevenue ? `${(metrics.pricing.totalRevenue / 1000000).toFixed(1)}M` : 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>ARPU</span>
                  <span className="font-semibold">{metrics?.pricing?.arpu ? `$${metrics.pricing.arpu}` : 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Price Elasticity</span>
                  <span className="font-semibold">{metrics?.pricing?.priceElasticity || 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Competitive Index</span>
                  <span className="font-semibold">{metrics?.pricing?.competitiveIndex || 'N/A'}</span>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Optimization</CardTitle>
                <CardDescription>Pricing optimization metrics</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between">
                  <span>Optimization ROI</span>
                  <span className="font-semibold text-green-600">{metrics?.pricing?.optimizationRoi ? `+${metrics.pricing.optimizationRoi}%` : 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Price Elasticity</span>
                  <span className="font-semibold">{metrics?.pricing?.priceElasticity || 'N/A'}</span>
                </div>
                <Button className="w-full mt-4">
                  Optimize Pricing
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
