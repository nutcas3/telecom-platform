import { SDKConfig, TelecomSDKConfig } from './config';
import { HttpClient } from './http-client';
import { SubscribersService } from './subscribers';
import { WebSocketClient, WebSocketEventHandler } from './websocket';
import { SystemStats, HealthStatus, Alert, RealTimeUsageUpdate, WebSocketMessage } from './types';

export class TelecomSDK {
  private static instance: TelecomSDK;
  private http: HttpClient;
  private ws: WebSocketClient;
  private subscribers: SubscribersService;

  private constructor() {
    this.http = new HttpClient();
    this.ws = new WebSocketClient();
    this.subscribers = new SubscribersService(this.http);
  }

  /**
   * Initialize the SDK with configuration
   */
  static initialize(config: TelecomSDKConfig): TelecomSDK {
    if (!TelecomSDK.instance) {
      SDKConfig.initialize(config);
      TelecomSDK.instance = new TelecomSDK();
    }
    return TelecomSDK.instance;
  }

  /**
   * Get the singleton instance
   */
  static getInstance(): TelecomSDK {
    if (!TelecomSDK.instance) {
      throw new Error('TelecomSDK not initialized. Call TelecomSDK.initialize() first.');
    }
    return TelecomSDK.instance;
  }

  /**
   * Subscribers service
   */
  get subscribersService(): SubscribersService {
    return this.subscribers;
  }

  /**
   * Connect to WebSocket for real-time updates
   */
  async connectWebSocket(): Promise<void> {
    await this.ws.connect();
  }

  /**
   * Disconnect from WebSocket
   */
  disconnectWebSocket(): void {
    this.ws.disconnect();
  }

  /**
   * Check if WebSocket is connected
   */
  get isWebSocketConnected(): boolean {
    return this.ws.isConnected;
  }

  /**
   * Subscribe to real-time usage updates
   */
  subscribeToUsage(imsi: string): void {
    this.ws.subscribeToUsage(imsi);
  }

  /**
   * Unsubscribe from usage updates
   */
  unsubscribeFromUsage(imsi: string): void {
    this.ws.unsubscribeFromUsage(imsi);
  }

  /**
   * Subscribe to alerts
   */
  subscribeToAlerts(): void {
    this.ws.subscribeToAlerts();
  }

  /**
   * Unsubscribe from alerts
   */
  unsubscribeFromAlerts(): void {
    this.ws.unsubscribeFromAlerts();
  }

  /**
   * Add WebSocket event handler
   */
  onWebSocketEvent(eventType: string, handler: WebSocketEventHandler): void {
    this.ws.on(eventType, handler);
  }

  /**
   * Remove WebSocket event handler
   */
  offWebSocketEvent(eventType: string, handler: WebSocketEventHandler): void {
    this.ws.off(eventType, handler);
  }

  /**
   * Get system statistics
   */
  async getSystemStats(): Promise<SystemStats> {
    return this.http.get<SystemStats>('/api/system/stats');
  }

  /**
   * Get system health status
   */
  async getHealthStatus(): Promise<HealthStatus> {
    return this.http.get<HealthStatus>('/api/system/health');
  }

  /**
   * Test API connection
   */
  async testConnection(): Promise<{ status: string; timestamp: string }> {
    return this.http.get('/api/test');
  }

  /**
   * Update API key
   */
  setApiKey(apiKey: string): void {
    this.http.setApiKey(apiKey);
  }

  /**
   * Update base URL
   */
  setBaseURL(baseURL: string): void {
    this.http.setBaseURL(baseURL);
  }

  /**
   * Enable/disable logging
   */
  setLogging(enabled: boolean): void {
    SDKConfig.getInstance().updateConfig({ enableLogging: enabled });
  }

  /**
   * Get current configuration
   */
  getConfig(): TelecomSDKConfig {
    const config = SDKConfig.getInstance();
    const result: TelecomSDKConfig = {
      baseURL: config.baseURL,
      timeout: config.timeout,
      retryAttempts: config.retryAttempts,
      retryDelay: config.retryDelay,
      enableLogging: config.enableLogging,
    };
    if (config.apiKey !== undefined) result.apiKey = config.apiKey;
    if (config.websocketURL !== undefined) result.websocketURL = config.websocketURL;
    return result;
  }

  /**
   * Convenience method for usage updates
   */
  onUsageUpdate(callback: (update: RealTimeUsageUpdate) => void): void {
    this.onWebSocketEvent('usage_update', (message: WebSocketMessage) => {
      callback(message.data as RealTimeUsageUpdate);
    });
  }

  /**
   * Convenience method for alerts
   */
  onAlert(callback: (alert: Alert) => void): void {
    this.onWebSocketEvent('alert', (message: WebSocketMessage) => {
      callback(message.data as Alert);
    });
  }

  /**
   * Cleanup resources
   */
  destroy(): void {
    this.disconnectWebSocket();
    // Clear any other resources if needed
  }
}

// Export types for external use
export * from './types';
export * from './config';
export { WebSocketClient } from './websocket';
export { HttpClient } from './http-client';
export { SubscribersService } from './subscribers';
