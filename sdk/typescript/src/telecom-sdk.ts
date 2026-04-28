import { AuthProvider } from './auth';
import { 
  HTTPClient, 
  SubscriberAPI, 
  UsageAPI, 
  PaymentAPI, 
  RatingPlanAPI, 
  SystemAPI, 
  GraphQLAPI 
} from './api';
import { WebSocketClient } from './websocket';
import { Alert, RealTimeUsageUpdate, WebSocketMessage } from './types';

export class TelecomSDK {
  private static instance: TelecomSDK;
  private authProvider: AuthProvider;
  private apiClient: HTTPClient;
  private ws: WebSocketClient;
  
  // API modules
  public subscribers: SubscriberAPI;
  public usage: UsageAPI;
  public payments: PaymentAPI;
  public ratingPlans: RatingPlanAPI;
  public system: SystemAPI;
  public graphql: GraphQLAPI;

  private constructor(config: {
    baseURL: string;
    apiKey?: string;
    jwtSecret?: string;
    timeout?: number;
    maxRetries?: number;
    retryDelay?: number;
    enableLogging?: boolean;
    websocketURL?: string;
  }) {
    const authConfig: any = {};
    if (config.apiKey !== undefined) authConfig.apiKey = config.apiKey;
    if (config.jwtSecret !== undefined) authConfig.jwtSecret = config.jwtSecret;
    
    this.authProvider = new AuthProvider(authConfig);
    
    this.apiClient = new HTTPClient(
      config.baseURL,
      this.authProvider,
      config.timeout || 30000,
      config.maxRetries || 3,
      config.retryDelay || 1000,
      config.enableLogging || false
    );
    
    this.ws = new WebSocketClient(
      config.websocketURL || config.baseURL.replace('http', 'ws'),
      config.enableLogging || false
    );
    
    // Initialize API modules
    this.subscribers = new SubscriberAPI(this.apiClient);
    this.usage = new UsageAPI(this.apiClient);
    this.payments = new PaymentAPI(this.apiClient);
    this.ratingPlans = new RatingPlanAPI(this.apiClient);
    this.system = new SystemAPI(this.apiClient);
    this.graphql = new GraphQLAPI(this.apiClient);
  }

  static initialize(config: {
    baseURL: string;
    apiKey?: string;
    jwtSecret?: string;
    timeout?: number;
    maxRetries?: number;
    retryDelay?: number;
    enableLogging?: boolean;
    websocketURL?: string;
  }): TelecomSDK {
    if (!TelecomSDK.instance) {
      TelecomSDK.instance = new TelecomSDK(config);
    }
    return TelecomSDK.instance;
  }

  static getInstance(): TelecomSDK {
    if (!TelecomSDK.instance) {
      throw new Error('TelecomSDK not initialized. Call TelecomSDK.initialize() first.');
    }
    return TelecomSDK.instance;
  }

  // Authentication methods
  generateJWTToken(userId: string, expiryHours: number = 24, additionalClaims: Record<string, any> = {}): string {
    return this.authProvider.generateJWTToken(userId, expiryHours, additionalClaims);
  }

  validateJWTToken(token: string): any {
    return this.authProvider.validateJWTToken(token);
  }

  // WebSocket methods
  async connectWebSocket(): Promise<void> {
    await this.ws.connect();
  }

  disconnectWebSocket(): void {
    this.ws.disconnect();
  }

  get isWebSocketConnected(): boolean {
    return this.ws.isConnected;
  }

  // Convenience methods
  onUsageUpdate(callback: (update: RealTimeUsageUpdate) => void): void {
    this.ws.on('usage_update', (message: WebSocketMessage) => {
      callback(message.data as RealTimeUsageUpdate);
    });
  }

  onAlert(callback: (alert: Alert) => void): void {
    this.ws.on('alert', (message: WebSocketMessage) => {
      callback(message.data as Alert);
    });
  }

  // Cleanup
  destroy(): void {
    this.disconnectWebSocket();
    this.apiClient.close();
  }
}

// Export types for external use
export * from './types';
export { AuthProvider, AuthConfig, JWTClaims } from './auth';
export * from './api';
export { WebSocketClient } from './websocket';
