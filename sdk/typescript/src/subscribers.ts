import { HttpClient } from './http-client';
import {
  Subscriber,
  CreateSubscriberRequest,
  UpdateSubscriberRequest,
  ListSubscribersRequest,
  PaginatedResponse,
  SubscriberAccount,
  TopUpRequest,
  Invoice,
  ProfileStatus,
} from './types';

export class SubscribersService {
  private http: HttpClient;

  constructor(http: HttpClient) {
    this.http = http;
  }

  /**
   * List all subscribers with pagination and filtering
   */
  async listSubscribers(request?: ListSubscribersRequest): Promise<PaginatedResponse<Subscriber>> {
    const params = {
      page: request?.page || 1,
      pageSize: request?.pageSize || 20,
      status: request?.status,
      organizationId: request?.organizationId,
      search: request?.search,
    };

    return this.http.get<PaginatedResponse<Subscriber>>('/api/subscribers', params);
  }

  /**
   * Get a specific subscriber by ID
   */
  async getSubscriber(id: number): Promise<Subscriber> {
    return this.http.get<Subscriber>(`/api/subscribers/${id}`);
  }

  /**
   * Get a subscriber by IMSI
   */
  async getSubscriberByImsi(imsi: string): Promise<Subscriber> {
    return this.http.get<Subscriber>(`/api/subscribers/imsi/${imsi}`);
  }

  /**
   * Create a new subscriber
   */
  async createSubscriber(request: CreateSubscriberRequest): Promise<Subscriber> {
    return this.http.post<Subscriber>('/api/subscribers', request);
  }

  /**
   * Update an existing subscriber
   */
  async updateSubscriber(id: number, request: UpdateSubscriberRequest): Promise<Subscriber> {
    return this.http.put<Subscriber>(`/api/subscribers/${id}`, request);
  }

  /**
   * Delete a subscriber
   */
  async deleteSubscriber(id: number): Promise<void> {
    await this.http.delete<void>(`/api/subscribers/${id}`);
  }

  /**
   * Get subscriber account information
   */
  async getSubscriberAccount(imsi: string): Promise<SubscriberAccount> {
    return this.http.get<SubscriberAccount>(`/api/subscribers/${imsi}/account`);
  }

  /**
   * Top up subscriber balance
   */
  async topUpBalance(imsi: string, request: TopUpRequest): Promise<SubscriberAccount> {
    return this.http.post<SubscriberAccount>(`/api/subscribers/${imsi}/top-up`, request);
  }

  /**
   * Suspend a subscriber
   */
  async suspendSubscriber(id: number): Promise<Subscriber> {
    return this.http.post<Subscriber>(`/api/subscribers/${id}/suspend`);
  }

  /**
   * Activate a subscriber
   */
  async activateSubscriber(id: number): Promise<Subscriber> {
    return this.http.post<Subscriber>(`/api/subscribers/${id}/activate`);
  }

  /**
   * Terminate a subscriber
   */
  async terminateSubscriber(id: number): Promise<Subscriber> {
    return this.http.post<Subscriber>(`/api/subscribers/${id}/terminate`);
  }

  /**
   * Provision eSIM profile for subscriber
   */
  async provisionESIM(imsi: string): Promise<{ profileId: string; activationCode: string }> {
    return this.http.post<{ profileId: string; activationCode: string }>(`/api/subscribers/${imsi}/esim/provision`);
  }

  /**
   * Activate eSIM profile
   */
  async activateESIM(imsi: string, profileId: string): Promise<void> {
    return this.http.post<void>(`/api/subscribers/${imsi}/esim/activate`, { profileId });
  }

  /**
   * Deactivate eSIM profile
   */
  async deactivateESIM(imsi: string, profileId: string): Promise<void> {
    return this.http.post<void>(`/api/subscribers/${imsi}/esim/deactivate`, { profileId });
  }

  /**
   * Get eSIM profile status
   */
  async getESIMStatus(imsi: string): Promise<ProfileStatus> {
    return this.http.get<ProfileStatus>(`/api/subscribers/${imsi}/esim/status`);
  }

  /**
   * Get subscriber invoices
   */
  async getInvoices(imsi: string, page = 1, pageSize = 20): Promise<PaginatedResponse<Invoice>> {
    return this.http.get<PaginatedResponse<Invoice>>(`/api/subscribers/${imsi}/invoices`, {
      page,
      pageSize,
    });
  }

  /**
   * Get specific invoice
   */
  async getInvoice(invoiceId: string): Promise<Invoice> {
    return this.http.get<Invoice>(`/api/invoices/${invoiceId}`);
  }

  /**
   * Download invoice PDF
   */
  async downloadInvoicePDF(invoiceId: string): Promise<Blob> {
    const response = await this.http.get(`/api/invoices/${invoiceId}/pdf`);
    return response as any; // Return blob for PDF download
  }

  /**
   * Get subscriber usage statistics
   */
  async getUsageStats(imsi: string, period: 'daily' | 'weekly' | 'monthly' = 'monthly'): Promise<{
    dataUsage: number;
    voiceUsage: number;
    smsUsage: number;
    cost: number;
    period: string;
  }> {
    return this.http.get(`/api/subscribers/${imsi}/usage/stats`, { period });
  }

  /**
   * Get real-time usage data
   */
  async getRealTimeUsage(imsi: string): Promise<{
    currentSession: {
      sessionId: string;
      startTime: string;
      dataUsed: number;
      voiceUsed: number;
      smsUsed: number;
      cost: number;
    } | null;
    todayUsage: {
      dataUsed: number;
      voiceUsed: number;
      smsUsed: number;
      cost: number;
    };
  }> {
    return this.http.get(`/api/subscribers/${imsi}/usage/realtime`);
  }
}
