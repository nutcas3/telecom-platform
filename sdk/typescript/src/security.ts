import { HTTPClient } from './api/http-client';
import { FraudAlert, FraudAlertFilter, FraudMetrics } from './types';

export class SecurityAPI {
  constructor(private client: HTTPClient) {}

  async analyzeTransaction(transaction: Record<string, any>): Promise<FraudAlert> {
    return this.client.request({
      method: 'POST',
      endpoint: '/api/v1/security/fraud/analyze',
      data: transaction
    });
  }

  async getFraudAlerts(filter: FraudAlertFilter): Promise<FraudAlert[]> {
    return this.client.request({
      method: 'POST',
      endpoint: '/api/v1/security/fraud/alerts',
      data: filter
    });
  }

  async updateAlertStatus(alertId: string, status: string, actions: string[]): Promise<void> {
    return this.client.request({
      method: 'PUT',
      endpoint: `/api/v1/security/fraud/alerts/${alertId}`,
      data: { status, actions }
    });
  }

  async getFraudMetrics(period: string): Promise<FraudMetrics> {
    return this.client.request({
      method: 'GET',
      endpoint: '/api/v1/security/fraud/metrics',
      params: { period }
    });
  }
}
