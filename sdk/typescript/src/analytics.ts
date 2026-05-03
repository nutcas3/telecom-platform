import { HTTPClient } from './api/http-client';
import {
  ChurnPrediction,
  ChurnMetrics,
  ChurnRiskLevel,
  MarketMetrics,
  PredictiveMaintenanceMetrics,
  PricingOptimizationResult,
  PricingMetrics
} from './types';

export class AnalyticsAPI {
  constructor(private client: HTTPClient) {}

  async predictChurn(profileId: string): Promise<ChurnPrediction> {
    return this.client.request({
      method: 'POST',
      endpoint: '/api/v1/analytics/churn/predict',
      data: { profileId }
    });
  }

  async getChurnMetrics(period: string): Promise<ChurnMetrics> {
    return this.client.request({
      method: 'GET',
      endpoint: '/api/v1/analytics/churn/metrics',
      params: { period }
    });
  }

  async getAtRiskCustomers(riskLevel: ChurnRiskLevel, limit: number): Promise<ChurnPrediction[]> {
    return this.client.request({
      method: 'GET',
      endpoint: '/api/v1/analytics/churn/at-risk',
      params: { riskLevel, limit: limit.toString() }
    });
  }

  async getMarketMetrics(period: string): Promise<MarketMetrics> {
    return this.client.request({
      method: 'GET',
      endpoint: '/api/v1/analytics/market/metrics',
      params: { period }
    });
  }

  async getPredictiveMaintenanceMetrics(period: string): Promise<PredictiveMaintenanceMetrics> {
    return this.client.request({
      method: 'GET',
      endpoint: '/api/v1/analytics/maintenance/metrics',
      params: { period }
    });
  }

  async getPricingMetrics(period: string): Promise<PricingMetrics> {
    return this.client.request({
      method: 'GET',
      endpoint: '/api/v1/analytics/pricing/metrics',
      params: { period }
    });
  }

  async optimizePrice(ratePlanId: string, strategy: string): Promise<PricingOptimizationResult> {
    return this.client.request({
      method: 'POST',
      endpoint: '/api/v1/analytics/pricing/optimize',
      data: { ratePlanId, strategy }
    });
  }
}
