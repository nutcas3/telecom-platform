export interface TelecomSDKConfig {
  baseURL: string;
  apiKey?: string;
  timeout?: number;
  retryAttempts?: number;
  retryDelay?: number;
  enableLogging?: boolean;
  websocketURL?: string;
}

export class SDKConfig {
  private static instance: SDKConfig;
  private config: TelecomSDKConfig;

  private constructor(config: TelecomSDKConfig) {
    this.config = {
      timeout: 30000,
      retryAttempts: 3,
      retryDelay: 1000,
      enableLogging: false,
      ...config
    };
  }

  public static initialize(config: TelecomSDKConfig): void {
    if (!SDKConfig.instance) {
      SDKConfig.instance = new SDKConfig(config);
    }
  }

  public static getInstance(): SDKConfig {
    if (!SDKConfig.instance) {
      throw new Error('SDKConfig not initialized. Call SDKConfig.initialize() first.');
    }
    return SDKConfig.instance;
  }

  public get baseURL(): string {
    return this.config.baseURL;
  }

  public get apiKey(): string | undefined {
    return this.config.apiKey;
  }

  public get timeout(): number {
    return this.config.timeout!;
  }

  public get retryAttempts(): number {
    return this.config.retryAttempts!;
  }

  public get retryDelay(): number {
    return this.config.retryDelay!;
  }

  public get enableLogging(): boolean {
    return this.config.enableLogging!;
  }

  public get websocketURL(): string | undefined {
    return this.config.websocketURL;
  }

  public updateConfig(updates: Partial<TelecomSDKConfig>): void {
    this.config = { ...this.config, ...updates };
  }
}
