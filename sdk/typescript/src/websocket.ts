import { SDKConfig } from './config';
import { WebSocketMessage } from './types';

export type WebSocketEventHandler = (message: WebSocketMessage) => void;

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private config: SDKConfig;
  private eventHandlers: Map<string, WebSocketEventHandler[]> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private isConnecting = false;
  private pingInterval: ReturnType<typeof setInterval> | null = null;

  constructor() {
    this.config = SDKConfig.getInstance();
  }

  /**
   * Connect to WebSocket server
   */
  async connect(): Promise<void> {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return;
    }

    if (!this.config.websocketURL) {
      throw new Error('WebSocket URL not configured');
    }

    this.isConnecting = true;

    try {
      this.ws = new WebSocket(this.config.websocketURL);

      await new Promise<void>((resolve, reject) => {
        if (!this.ws) return reject(new Error('WebSocket not initialized'));

        this.ws.onopen = () => {
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          this.startPing();
          if (this.config.enableLogging) {
            console.log('[SDK] WebSocket connected');
          }
          resolve();
        };

        this.ws.onmessage = (event: MessageEvent) => {
          try {
            const message: WebSocketMessage = JSON.parse(
              typeof event.data === 'string' ? event.data : new TextDecoder().decode(event.data),
            );
            this.handleMessage(message);
          } catch (error) {
            if (this.config.enableLogging) {
              console.error('[SDK] Failed to parse WebSocket message:', error);
            }
          }
        };

        this.ws.onclose = (event: CloseEvent) => {
          this.isConnecting = false;
          this.stopPing();
          if (this.config.enableLogging) {
            console.log(`[SDK] WebSocket closed: ${event.code} - ${event.reason}`);
          }
          if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
        };

        this.ws.onerror = () => {
          this.isConnecting = false;
          if (this.config.enableLogging) {
            console.error('[SDK] WebSocket error');
          }
          reject(new Error('WebSocket connection failed'));
        };
      });
    } catch (error) {
      this.isConnecting = false;
      throw error;
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    this.stopPing();
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
    this.reconnectAttempts = this.maxReconnectAttempts;
  }

  /**
   * Send message to WebSocket server
   */
  send(message: unknown): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected');
    }
    this.ws.send(JSON.stringify(message));
  }

  /**
   * Subscribe to real-time usage updates for a subscriber
   */
  subscribeToUsage(imsi: string): void {
    this.send({ type: 'subscribe', channel: 'usage', imsi });
  }

  /**
   * Unsubscribe from usage updates
   */
  unsubscribeFromUsage(imsi: string): void {
    this.send({ type: 'unsubscribe', channel: 'usage', imsi });
  }

  /**
   * Subscribe to alerts
   */
  subscribeToAlerts(): void {
    this.send({ type: 'subscribe', channel: 'alerts' });
  }

  /**
   * Unsubscribe from alerts
   */
  unsubscribeFromAlerts(): void {
    this.send({ type: 'unsubscribe', channel: 'alerts' });
  }

  /**
   * Add event handler for specific message type
   */
  on(eventType: string, handler: WebSocketEventHandler): void {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, []);
    }
    this.eventHandlers.get(eventType)!.push(handler);
  }

  /**
   * Remove event handler
   */
  off(eventType: string, handler: WebSocketEventHandler): void {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  /**
   * Get connection status
   */
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  private handleMessage(message: WebSocketMessage): void {
    const handlers = this.eventHandlers.get(message.type);
    if (handlers) {
      handlers.forEach(handler => handler(message));
    }
  }

  private scheduleReconnect(): void {
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    setTimeout(() => {
      this.reconnectAttempts++;
      if (this.config.enableLogging) {
        console.log(`[SDK] Reconnecting (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      }
      this.connect().catch(error => {
        if (this.config.enableLogging) {
          console.error('[SDK] Reconnection failed:', error);
        }
      });
    }, delay);
  }

  private startPing(): void {
    this.pingInterval = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping', timestamp: new Date().toISOString() });
      }
    }, 30000);
  }

  private stopPing(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }
}
