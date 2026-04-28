import { HTTPClient } from './http-client';
import { UsageEvent, RealTimeUsageUpdate, PaginatedResponse, SubscriberAccount } from '../types';

export class UsageAPI {
  constructor(private client: HTTPClient) {}

  async getStats(
    subscriberId: number,
    startDate: string,
    endDate: string
  ): Promise<SubscriberAccount> {
    const params = {
      start_date: startDate,
      end_date: endDate,
    };

    return this.client.request<SubscriberAccount>({
      method: 'GET',
      endpoint: `/v1/subscribers/${subscriberId}/usage`,
      params,
    });
  }

  async listEvents(
    subscriberId?: number,
    usageType?: string,
    startDate?: string,
    endDate?: string,
    page: number = 1,
    pageSize: number = 50
  ): Promise<PaginatedResponse<UsageEvent>> {
    const params: Record<string, string> = {
      page: page.toString(),
      page_size: pageSize.toString(),
    };

    if (subscriberId) {
      params.subscriber_id = subscriberId.toString();
    }
    if (usageType) {
      params.usage_type = usageType;
    }
    if (startDate) {
      params.start_date = startDate;
    }
    if (endDate) {
      params.end_date = endDate;
    }

    return this.client.request<PaginatedResponse<UsageEvent>>({
      method: 'GET',
      endpoint: '/v1/usage/events',
      params,
    });
  }

  async getRealtime(subscriberId: number): Promise<RealTimeUsageUpdate> {
    return this.client.request<RealTimeUsageUpdate>({
      method: 'GET',
      endpoint: `/v1/subscribers/${subscriberId}/realtime`,
    });
  }
}
