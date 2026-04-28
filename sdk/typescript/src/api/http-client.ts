import { AuthProvider } from '../auth';

export interface RequestConfig {
  method: string;
  endpoint: string;
  data?: any;
  params?: Record<string, string>;
}

export class HTTPClient {
  private baseURL: string;
  private authProvider: AuthProvider;
  private timeout: number;
  private maxRetries: number;
  private retryDelay: number;
  private enableLogging: boolean;

  constructor(
    baseURL: string,
    authProvider: AuthProvider,
    timeout: number = 30000,
    maxRetries: number = 3,
    retryDelay: number = 1000,
    enableLogging: boolean = false
  ) {
    this.baseURL = baseURL.replace(/\/$/, '');
    this.authProvider = authProvider;
    this.timeout = timeout;
    this.maxRetries = maxRetries;
    this.retryDelay = retryDelay;
    this.enableLogging = enableLogging;
  }

  async request<T>(config: RequestConfig): Promise<T> {
    const { method, endpoint, data, params } = config;
    const url = new URL(`${this.baseURL}${endpoint}`);

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, value);
      });
    }

    const headers = this.authProvider.getHeaders();
    const options: RequestInit = {
      method,
      headers,
    };

    if (data && method !== 'GET') {
      options.body = JSON.stringify(data);
    }

    let lastError: Error | null = null;

    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        const response = await fetch(url.toString(), {
          ...options,
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        await this.handleResponseErrors(response);

        return await response.json();
      } catch (error) {
        lastError = error as Error;
        
        if (attempt === this.maxRetries) {
          throw new Error(`Request failed after ${this.maxRetries} retries: ${lastError.message}`);
        }

        if (this.enableLogging) {
          console.log(`Request failed (attempt ${attempt + 1}), retrying in ${this.retryDelay}ms: ${lastError.message}`);
        }

        await this.delay(this.retryDelay * Math.pow(2, attempt));
      }
    }

    throw lastError;
  }

  private async handleResponseErrors(response: Response): Promise<void> {
    if (response.status === 401) {
      throw new Error('Authentication failed');
    } else if (response.status === 429) {
      throw new Error('Rate limit exceeded');
    } else if (response.status >= 400 && response.status < 500) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(`API error: ${errorData.error || 'Bad request'}`);
    } else if (response.status >= 500) {
      throw new Error(`Server error: ${response.status}`);
    }
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  close(): void {
    // Cleanup if needed
  }
}
