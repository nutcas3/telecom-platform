const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

export interface ChurnMetrics {
  totalSubscribers: number;
  churnedSubscribers: number;
  churnRate: number;
  monthlyChurnRate: number;
  annualChurnRate: number;
  averageTenureDays: number;
  riskDistribution: {
    low: number;
    medium: number;
    high: number;
    critical: number;
  };
}

export interface ChurnPrediction {
  profileId: string;
  riskLevel: string;
  riskScore: number;
  predictedChurnDate?: string;
  reasons: string[];
  recommendations: string[];
  lastUpdated: string;
}

export interface FraudAlert {
  id: string;
  type: string;
  severity: string;
  profileId: string;
  description: string;
  riskScore: number;
  evidence: string[];
  ipAddress: string;
  timestamp: string;
  status: string;
  actionsTaken: string[];
}

export interface FraudMetrics {
  totalAlerts: number;
  resolvedAlerts: number;
  falsePositives: number;
  resolutionRate: number;
  falsePositiveRate: number;
  byType: Record<string, number>;
  bySeverity: Record<string, number>;
}

export interface MarketMetrics {
  totalMarketSize: number;
  ourSubscribers: number;
  marketShare: number;
  growthRate: number;
  byCountry: Record<string, {
    marketSize: number;
    ourSubs: number;
    penetration: number;
  }>;
}

export interface PricingMetrics {
  totalRevenue: number;
  arpu: number;
  priceElasticity: number;
  competitiveIndex: number;
  optimizationRoi: number;
  byPlan: Record<string, number>;
  byRegion: Record<string, number>;
}

export interface PricingOptimizationResult {
  ratePlanId: string;
  strategy: string;
  currentPrice: number;
  optimalPrice: number;
  priceChangePct: number;
  expectedRevenue: number;
  expectedDemand: number;
  confidence: number;
  reasoning: string[];
  risks: string[];
  recommendations: string[];
}

export interface MaintenanceMetrics {
  totalAssets: number;
  healthyAssets: number;
  assetsNeedingAttention: number;
  uptime: number;
  meanTimeToFailure: number;
  meanTimeToRepair: number;
  byType: Record<string, number>;
  byStatus: Record<string, number>;
}

export interface Asset {
  id: string;
  name: string;
  type: string;
  health: number;
  status: string;
  lastMaintenance: string;
  nextMaintenance?: string;
  predictedFailure?: string;
  riskFactors: string[];
}

class ApiClient {
  private baseUrl: string;
  private headers: Record<string, string>;

  constructor() {
    this.baseUrl = API_BASE_URL;
    this.headers = {
      'Content-Type': 'application/json',
    };

    // Add auth token if available
    const token = this.getAuthToken();
    if (token) {
      this.headers['Authorization'] = `Bearer ${token}`;
    }
  }

  private getAuthToken(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('auth_token');
    }
    return null;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    try {
      const url = `${this.baseUrl}${endpoint}`;
      const response = await fetch(url, {
        ...options,
        headers: {
          ...this.headers,
          ...options.headers,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return { data };
    } catch (error) {
      console.error('API request failed:', error);
      return { error: error instanceof Error ? error.message : 'Unknown error' };
    }
  }

  // Analytics API
  async getChurnMetrics(period: string = 'monthly'): Promise<ApiResponse<ChurnMetrics>> {
    return this.request<ChurnMetrics>(`/api/v1/analytics/churn/metrics?period=${period}`);
  }

  async getChurnPredictions(riskLevel?: string, limit?: number): Promise<ApiResponse<ChurnPrediction[]>> {
    const params = new URLSearchParams();
    if (riskLevel) params.append('risk_level', riskLevel);
    if (limit) params.append('limit', limit.toString());
    
    return this.request<ChurnPrediction[]>(`/api/v1/analytics/churn/predictions?${params}`);
  }

  async predictChurn(profileId: string): Promise<ApiResponse<ChurnPrediction>> {
    return this.request<ChurnPrediction>('/api/v1/analytics/churn/predict', {
      method: 'POST',
      body: JSON.stringify({ profile_id: profileId }),
    });
  }

  // Fraud Detection API
  async getFraudAlerts(severity?: string): Promise<ApiResponse<FraudAlert[]>> {
    const params = severity ? `?severity=${severity}` : '';
    return this.request<FraudAlert[]>(`/api/v1/security/fraud/alerts${params}`);
  }

  async getFraudMetrics(period: string = 'monthly'): Promise<ApiResponse<FraudMetrics>> {
    return this.request<FraudMetrics>(`/api/v1/security/fraud/metrics?period=${period}`);
  }

  async analyzeTransaction(transaction: Record<string, any>): Promise<ApiResponse<FraudAlert>> {
    return this.request<FraudAlert>('/api/v1/security/fraud/analyze', {
      method: 'POST',
      body: JSON.stringify(transaction),
    });
  }

  async updateFraudAlert(alertId: string, status: string, actions: string[]): Promise<ApiResponse<void>> {
    return this.request<void>(`/api/v1/security/fraud/alerts/${alertId}`, {
      method: 'PUT',
      body: JSON.stringify({ status, actions }),
    });
  }

  // Market Analytics API
  async getMarketMetrics(period: string = 'monthly'): Promise<ApiResponse<MarketMetrics>> {
    return this.request<MarketMetrics>(`/api/v1/analytics/market/metrics?period=${period}`);
  }

  async getCompetitors(): Promise<ApiResponse<Record<string, any>[]>> {
    return this.request<Record<string, any>[]>('/api/v1/analytics/market/competitors');
  }

  async getMarketOpportunities(): Promise<ApiResponse<Record<string, any>[]>> {
    return this.request<Record<string, any>[]>('/api/v1/analytics/market/opportunities');
  }

  // Pricing API
  async getPricingMetrics(period: string = 'monthly'): Promise<ApiResponse<PricingMetrics>> {
    return this.request<PricingMetrics>(`/api/v1/analytics/pricing/metrics?period=${period}`);
  }

  async optimizePricing(ratePlanIds: string[], strategy: string): Promise<ApiResponse<PricingOptimizationResult[]>> {
    return this.request<PricingOptimizationResult[]>('/api/v1/analytics/pricing/optimize', {
      method: 'POST',
      body: JSON.stringify({ rate_plan_ids: ratePlanIds, strategy }),
    });
  }

  async getPriceElasticity(): Promise<ApiResponse<Record<string, number>>> {
    return this.request<Record<string, number>>('/api/v1/analytics/pricing/elasticity');
  }

  // Maintenance API
  async getMaintenanceMetrics(period: string = 'monthly'): Promise<ApiResponse<MaintenanceMetrics>> {
    return this.request<MaintenanceMetrics>(`/api/v1/analytics/maintenance/metrics?period=${period}`);
  }

  async getAssets(): Promise<ApiResponse<Asset[]>> {
    return this.request<Asset[]>('/api/v1/analytics/maintenance/assets');
  }

  async getMaintenanceAlerts(): Promise<ApiResponse<Record<string, any>[]>> {
    return this.request<Record<string, any>[]>('/api/v1/analytics/maintenance/alerts');
  }

  async predictFailure(assetId: string): Promise<ApiResponse<Record<string, any>>> {
    return this.request<Record<string, any>>(`/api/v1/analytics/maintenance/predict/${assetId}`, {
      method: 'POST',
    });
  }
}

export const apiClient = new ApiClient();
