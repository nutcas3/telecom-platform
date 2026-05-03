"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { TrendingUp, TrendingDown, DollarSign, Target, Zap, BarChart3, AlertTriangle } from "lucide-react";
import { apiClient, PricingMetrics, PricingOptimizationResult } from "@/lib/api-client";

export default function PricingPage() {
  const [metrics, setMetrics] = useState<PricingMetrics | null>(null);
  const [optimizations, setOptimizations] = useState<PricingOptimizationResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedStrategy, setSelectedStrategy] = useState<string>("revenue_maximization");

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        const metricsResponse = await apiClient.getPricingMetrics();

        if (metricsResponse.error) {
          setError('Failed to fetch pricing metrics');
        }

        setMetrics(metricsResponse.data || null);
      } catch (err) {
        console.error('Failed to fetch pricing data:', err);
        setError('Failed to load pricing optimization data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const handleOptimizePricing = async () => {
    try {
      setLoading(true);
      const response = await apiClient.optimizePricing(['plan-1', 'plan-2', 'plan-3'], selectedStrategy);
      
      if (response.error) {
        setError('Failed to optimize pricing');
      } else {
        setOptimizations(response.data || []);
      }
    } catch (err) {
      console.error('Failed to optimize pricing:', err);
      setError('Failed to optimize pricing');
    } finally {
      setLoading(false);
    }
  };

  const getStrategyBadgeVariant = (strategy: string) => {
    switch (strategy) {
      case "revenue_maximization": return "default";
      case "market_share": return "secondary";
      case "profit_margin": return "outline";
      case "competitive": return "destructive";
      default: return "secondary";
    }
  };

  const getChangeIcon = (change: number) => {
    return change >= 0 ? <TrendingUp className="h-4 w-4" /> : <TrendingDown className="h-4 w-4" />;
  };

  const getChangeColor = (change: number) => {
    return change >= 0 ? "text-green-600" : "text-red-600";
  };

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 85) return "text-green-600";
    if (confidence >= 70) return "text-yellow-600";
    return "text-red-600";
  };

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
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Error Loading Pricing Optimization</h3>
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
        <div>
          <h1 className="text-3xl font-bold">Pricing Optimization</h1>
          <p className="text-muted-foreground">Optimize pricing strategies for maximum revenue and market share</p>
        </div>
        <div className="flex items-center space-x-4">
          <select
            value={selectedStrategy}
            onChange={(e) => setSelectedStrategy(e.target.value)}
            className="px-3 py-2 border rounded-md"
          >
            <option value="revenue_maximization">Revenue Maximization</option>
            <option value="market_share">Market Share</option>
            <option value="profit_margin">Profit Margin</option>
            <option value="competitive">Competitive</option>
          </select>
          <Button onClick={handleOptimizePricing} disabled={loading}>
            {loading ? "Optimizing..." : "Optimize Pricing"}
          </Button>
        </div>
      </div>

      {/* Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${metrics?.totalRevenue ? (metrics.totalRevenue / 1000000).toFixed(1) : '0'}M</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">ARPU</CardTitle>
            <Target className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${metrics?.arpu || '0'}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Price Elasticity</CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.priceElasticity || '0'}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Optimization ROI</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">+{metrics?.optimizationRoi || '0'}%</div>
          </CardContent>
        </Card>
      </div>

      {/* Optimization Results */}
      {optimizations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Optimization Results</CardTitle>
            <CardDescription>Recommended pricing adjustments based on {selectedStrategy} strategy</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Rate Plan</TableHead>
                  <TableHead>Current Price</TableHead>
                  <TableHead>Optimal Price</TableHead>
                  <TableHead>Change</TableHead>
                  <TableHead>Expected Revenue</TableHead>
                  <TableHead>Confidence</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {optimizations.map((opt, index) => (
                  <TableRow key={index}>
                    <TableCell className="font-medium">{opt.ratePlanId}</TableCell>
                    <TableCell>${opt.currentPrice.toFixed(2)}</TableCell>
                    <TableCell>${opt.optimalPrice.toFixed(2)}</TableCell>
                    <TableCell>
                      <div className="flex items-center space-x-2">
                        {getChangeIcon(opt.priceChangePct)}
                        <span className={getChangeColor(opt.priceChangePct)}>
                          {opt.priceChangePct > 0 ? '+' : ''}{opt.priceChangePct.toFixed(1)}%
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>${(opt.expectedRevenue / 1000000).toFixed(2)}M</TableCell>
                    <TableCell>
                      <Badge variant={opt.confidence >= 80 ? "default" : "secondary"}>
                        {opt.confidence.toFixed(0)}%
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
