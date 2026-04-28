import { HTTPClient } from './http-client';
import { Invoice, PaymentMethod, PaginatedResponse } from '../types';

export class PaymentAPI {
  constructor(private client: HTTPClient) {}

  async createTransaction(data: {
    subscriberId: number;
    amount: number;
    currency?: string;
    gateway?: string;
    metadata?: Record<string, any>;
  }): Promise<any> {
    return this.client.request<any>({
      method: 'POST',
      endpoint: '/v1/payments/transactions',
      data,
    });
  }

  async getTransaction(transactionId: string): Promise<any> {
    return this.client.request<any>({
      method: 'GET',
      endpoint: `/v1/payments/transactions/${transactionId}`,
    });
  }

  async listTransactions(
    subscriberId?: number,
    status?: string,
    page: number = 1,
    pageSize: number = 50
  ): Promise<PaginatedResponse<any>> {
    const params: Record<string, string> = {
      page: page.toString(),
      page_size: pageSize.toString(),
    };

    if (subscriberId) {
      params.subscriber_id = subscriberId.toString();
    }
    if (status) {
      params.status = status;
    }

    return this.client.request<PaginatedResponse<any>>({
      method: 'GET',
      endpoint: '/v1/payments/transactions',
      params,
    });
  }

  async listInvoices(
    subscriberId?: number,
    status?: string,
    page: number = 1,
    pageSize: number = 50
  ): Promise<PaginatedResponse<Invoice>> {
    const params: Record<string, string> = {
      page: page.toString(),
      page_size: pageSize.toString(),
    };

    if (subscriberId) {
      params.subscriber_id = subscriberId.toString();
    }
    if (status) {
      params.status = status;
    }

    return this.client.request<PaginatedResponse<Invoice>>({
      method: 'GET',
      endpoint: '/v1/payments/invoices',
      params,
    });
  }

  async addPaymentMethod(data: {
    subscriberId: number;
    type: string;
    token: string;
    isDefault?: boolean;
  }): Promise<PaymentMethod> {
    return this.client.request<PaymentMethod>({
      method: 'POST',
      endpoint: '/v1/payments/methods',
      data,
    });
  }
}
