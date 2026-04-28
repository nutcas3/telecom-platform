import { HTTPClient } from './http-client';
import { SystemStats, HealthStatus } from '../types';

export class SystemAPI {
  constructor(private client: HTTPClient) {}

  async getStats(): Promise<SystemStats> {
    return this.client.request<SystemStats>({
      method: 'GET',
      endpoint: '/v1/system/stats',
    });
  }

  async getHealth(): Promise<HealthStatus> {
    return this.client.request<HealthStatus>({
      method: 'GET',
      endpoint: '/v1/health',
    });
  }
}
