import { HTTPClient } from './api/http-client';

export interface ConvertRequest {
  from: string;
  to: string;
  amount: number;
}

export interface ConvertResponse {
  from: string;
  to: string;
  amount: number;
  converted: number;
  rate: number;
  timestamp: string;
}

export interface ExchangeRate {
  from: string;
  to: string;
  rate: number;
  timestamp: string;
}

export interface BillingTransaction {
  id: string;
  profileId: string;
  amount: number;
  currency: string;
  type: string;
  status: string;
  description: string;
  createdAt: string;
}

export interface BillingSummary {
  profileId: string;
  period: string;
  totalAmount: number;
  currency: string;
  transactionCount: number;
  breakdown: Record<string, number>;
}

export class CurrencyAPI {
  constructor(private client: HTTPClient) {}

  async convert(request: ConvertRequest): Promise<ConvertResponse> {
    return this.client.request({
      method: 'POST',
      endpoint: '/api/v1/currency/convert',
      data: request
    });
  }

  async getExchangeRate(from: string, to: string): Promise<ExchangeRate> {
    return this.client.request({
      method: 'GET',
      endpoint: `/api/v1/currency/exchange/${from}/${to}`
    });
  }

  async getExchangeRateHistory(from: string, to: string, days: number = 30): Promise<ExchangeRate[]> {
    return this.client.request({
      method: 'GET',
      endpoint: `/api/v1/currency/exchange/${from}/${to}/history`,
      params: { days: days.toString() }
    });
  }

  async getSupportedCurrencies(): Promise<string[]> {
    return this.client.request({
      method: 'GET',
      endpoint: '/api/v1/currency/currencies'
    });
  }

  async refreshExchangeRates(): Promise<void> {
    return this.client.request({
      method: 'POST',
      endpoint: '/api/v1/currency/exchange/refresh',
      data: {}
    });
  }

  async processBilling(billingData: Record<string, any>): Promise<BillingTransaction> {
    return this.client.request({
      method: 'POST',
      endpoint: '/api/v1/currency/billing',
      data: billingData
    });
  }

  async getBillingHistory(profileId: string, limit: number = 50): Promise<BillingTransaction[]> {
    return this.client.request({
      method: 'GET',
      endpoint: `/api/v1/currency/billing/history/${profileId}`,
      params: { limit: limit.toString() }
    });
  }

  async getBillingSummary(profileId: string, period: string = 'monthly'): Promise<BillingSummary> {
    return this.client.request({
      method: 'GET',
      endpoint: `/api/v1/currency/billing/summary/${profileId}`,
      params: { period }
    });
  }

  async processRefund(transactionId: string, reason: string): Promise<BillingTransaction> {
    return this.client.request({
      method: 'POST',
      endpoint: `/api/v1/currency/billing/refund/${transactionId}`,
      data: { reason }
    });
  }

  async getBillingAnalytics(period: string = 'monthly'): Promise<Record<string, any>> {
    return this.client.request({
      method: 'GET',
      endpoint: '/api/v1/currency/billing/analytics',
      params: { period }
    });
  }
}
