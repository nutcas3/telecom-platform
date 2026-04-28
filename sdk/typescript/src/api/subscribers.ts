import { HTTPClient } from './http-client';
import { Subscriber, PaginatedResponse } from '../types';

export class SubscriberAPI {
  constructor(private client: HTTPClient) {}

  async get(subscriberId: number): Promise<Subscriber> {
    return this.client.request<Subscriber>({
      method: 'GET',
      endpoint: `/v1/subscribers/${subscriberId}`,
    });
  }

  async list(
    page: number = 1,
    pageSize: number = 50,
    status?: string
  ): Promise<PaginatedResponse<Subscriber>> {
    const params: Record<string, string> = {
      page: page.toString(),
      page_size: pageSize.toString(),
    };

    if (status) {
      params.status = status;
    }

    return this.client.request<PaginatedResponse<Subscriber>>({
      method: 'GET',
      endpoint: '/v1/subscribers',
      params,
    });
  }

  async create(data: {
    imsi: string;
    msisdn: string;
    firstName: string;
    lastName: string;
    email: string;
    planId: number;
    organizationId?: string;
  }): Promise<Subscriber> {
    return this.client.request<Subscriber>({
      method: 'POST',
      endpoint: '/v1/subscribers',
      data,
    });
  }

  async update(subscriberId: number, data: Record<string, any>): Promise<Subscriber> {
    return this.client.request<Subscriber>({
      method: 'PUT',
      endpoint: `/v1/subscribers/${subscriberId}`,
      data,
    });
  }

  async suspend(subscriberId: number): Promise<Subscriber> {
    return this.client.request<Subscriber>({
      method: 'POST',
      endpoint: `/v1/subscribers/${subscriberId}/suspend`,
    });
  }

  async activate(subscriberId: number): Promise<Subscriber> {
    return this.client.request<Subscriber>({
      method: 'POST',
      endpoint: `/v1/subscribers/${subscriberId}/activate`,
    });
  }

  async terminate(subscriberId: number): Promise<{ success: boolean }> {
    return this.client.request<{ success: boolean }>({
      method: 'DELETE',
      endpoint: `/v1/subscribers/${subscriberId}`,
    });
  }
}
