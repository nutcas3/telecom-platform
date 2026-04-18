import { Error as SdkError } from './types';
import { SDKConfig } from './config';

export class HttpClient {
  private config: SDKConfig;

  constructor() {
    this.config = SDKConfig.getInstance();
  }

  private get headers(): Record<string, string> {
    const h: Record<string, string> = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    };
    if (this.config.apiKey) {
      h['Authorization'] = `Bearer ${this.config.apiKey}`;
    }
    return h;
  }

  private buildURL(path: string, params?: Record<string, unknown>): string {
    const url = new URL(path, this.config.baseURL);
    if (params) {
      Object.entries(params).forEach(([k, v]) => {
        if (v !== undefined && v !== null) {
          url.searchParams.set(k, String(v));
        }
      });
    }
    return url.toString();
  }

  private async fetchWithRetry(url: string, init: RequestInit, attempt = 0): Promise<Response> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.config.timeout);
    try {
      if (this.config.enableLogging) {
        console.log(`[SDK] ${init.method ?? 'GET'} ${url}`);
      }

      const response = await fetch(url, { ...init, signal: controller.signal });

      if (this.config.enableLogging) {
        console.log(`[SDK] Response ${response.status}`);
      }

      if (!response.ok && this.shouldRetry(response.status) && attempt < this.config.retryAttempts) {
        await new Promise(r => setTimeout(r, this.config.retryDelay * Math.pow(2, attempt)));
        return this.fetchWithRetry(url, init, attempt + 1);
      }

      return response;
    } catch (err) {
      if (attempt < this.config.retryAttempts && !(err instanceof DOMException)) {
        await new Promise(r => setTimeout(r, this.config.retryDelay * Math.pow(2, attempt)));
        return this.fetchWithRetry(url, init, attempt + 1);
      }
      throw err;
    } finally {
      clearTimeout(timeout);
    }
  }

  private shouldRetry(status: number): boolean {
    return status >= 500 || status === 429;
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      let body: any = {};
      try { body = await response.json(); } catch { /* empty */ }
      const err: SdkError = {
        code: body.code ?? `HTTP_${response.status}`,
        message: body.message ?? response.statusText,
        details: body.details,
        timestamp: new Date().toISOString(),
      };
      throw err;
    }
    return response.json() as Promise<T>;
  }

  async get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
    const response = await this.fetchWithRetry(
      this.buildURL(url, params),
      { method: 'GET', headers: this.headers },
    );
    return this.handleResponse<T>(response);
  }

  async post<T>(url: string, data?: unknown): Promise<T> {
    const response = await this.fetchWithRetry(
      this.buildURL(url),
      { method: 'POST', headers: this.headers, body: data ? JSON.stringify(data) : null },
    );
    return this.handleResponse<T>(response);
  }

  async put<T>(url: string, data?: unknown): Promise<T> {
    const response = await this.fetchWithRetry(
      this.buildURL(url),
      { method: 'PUT', headers: this.headers, body: data ? JSON.stringify(data) : null },
    );
    return this.handleResponse<T>(response);
  }

  async patch<T>(url: string, data?: unknown): Promise<T> {
    const response = await this.fetchWithRetry(
      this.buildURL(url),
      { method: 'PATCH', headers: this.headers, body: data ? JSON.stringify(data) : null },
    );
    return this.handleResponse<T>(response);
  }

  async delete<T>(url: string): Promise<T> {
    const response = await this.fetchWithRetry(
      this.buildURL(url),
      { method: 'DELETE', headers: this.headers },
    );
    return this.handleResponse<T>(response);
  }

  setApiKey(apiKey: string): void {
    this.config.updateConfig({ apiKey });
  }

  setBaseURL(baseURL: string): void {
    this.config.updateConfig({ baseURL });
  }
}
