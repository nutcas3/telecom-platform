import { HTTPClient } from './http-client';
import { RatingPlan } from '../types';

export class RatingPlanAPI {
  constructor(private client: HTTPClient) {}

  async list(): Promise<RatingPlan[]> {
    return this.client.request<RatingPlan[]>({
      method: 'GET',
      endpoint: '/v1/rating-plans',
    });
  }

  async get(planId: string): Promise<RatingPlan> {
    return this.client.request<RatingPlan>({
      method: 'GET',
      endpoint: `/v1/rating-plans/${planId}`,
    });
  }
}
