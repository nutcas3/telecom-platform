export interface Subscriber {
  id: number;
  imsi: string;
  msisdn: string;
  firstName: string;
  lastName: string;
  email: string;
  organizationId?: string;
  planId: number;
  euiccId?: string;
  profileId?: string;
  profileStatus: ProfileStatus;
  status: SubscriberStatus;
  balance: number;
  dataLimit: number;
  dataUsed: number;
  voiceLimit: number;
  voiceUsed: number;
  smsLimit: number;
  smsUsed: number;
  createdAt: string;
  updatedAt: string;
}

export enum SubscriberStatus {
  Active = 'active',
  Inactive = 'inactive',
  Suspended = 'suspended',
  Terminated = 'terminated',
  Provisioning = 'provisioning'
}

export enum ProfileStatus {
  Active = 'active',
  Inactive = 'inactive',
  Downloading = 'downloading',
  Failed = 'failed'
}

export interface CreateSubscriberRequest {
  msisdn: string;
  firstName: string;
  lastName: string;
  email: string;
  organizationId?: string;
  planId: number;
  euiccId?: string;
}

// Churn Analysis Types
export enum ChurnRiskLevel {
  Low = 'low',
  Medium = 'medium',
  High = 'high',
  Critical = 'critical'
}

export interface ChurnPrediction {
  profileId: string;
  riskLevel: ChurnRiskLevel;
  riskScore: number;
  predictedChurnDate?: string;
  reasons: string[];
  recommendations: string[];
  lastUpdated: string;
}

export interface ChurnMetrics {
  period: string;
  totalSubscribers: number;
  churnedSubscribers: number;
  churnRate: number;
  monthlyChurnRate: number;
  annualChurnRate: number;
  averageTenureDays: number;
  riskDistribution: Record<ChurnRiskLevel, number>;
  generatedAt: string;
}

// Fraud Detection Types
export enum FraudType {
  AccountTakeover = 'account_takeover',
  SubscriptionFraud = 'subscription_fraud',
  PaymentFraud = 'payment_fraud',
  UsageAnomaly = 'usage_anomaly',
  SIMSwap = 'sim_swap'
}

export enum FraudSeverity {
  Low = 'low',
  Medium = 'medium',
  High = 'high',
  Critical = 'critical'
}

export interface FraudAlert {
  id: string;
  type: FraudType;
  severity: FraudSeverity;
  profileId: string;
  description: string;
  riskScore: number;
  evidence: string[];
  ipAddress: string;
  timestamp: string;
  status: string;
  actionsTaken: string[];
  metadata: Record<string, any>;
}

export interface FraudMetrics {
  period: string;
  totalAlerts: number;
  resolvedAlerts: number;
  falsePositives: number;
  resolutionRatePct: number;
  falsePositiveRatePct: number;
  byType: Record<FraudType, number>;
  bySeverity: Record<FraudSeverity, number>;
  generatedAt: string;
}

export interface FraudAlertFilter {
  type?: FraudType;
  severity?: FraudSeverity;
  status?: string;
  fromDate?: string;
  toDate?: string;
  limit?: number;
}

// Market Analytics Types
export interface MarketMetrics {
  period: string;
  totalMarketSize: number;
  ourSubscribers: number;
  marketSharePct: number;
  growthRatePct: number;
  byCountry: Record<string, CountryMetrics>;
  byCarrier: Record<string, MarketCarrierMetrics>;
  byDemographic: Record<string, DemoMetrics>;
  competitorAnalysis: Record<string, CompetitorMetrics>;
  marketOpportunities: MarketOpportunity[];
  generatedAt: string;
}

export interface CountryMetrics {
  country: string;
  marketSize: number;
  ourSubscribers: number;
  marketSharePct: number;
  growthRatePct: number;
  averageRevenue: number;
}

export interface MarketCarrierMetrics {
  carrierId: string;
  carrierName: string;
  subscribers: number;
  marketSharePct: number;
  averageRevenue: number;
  qualityScore: number;
}

export interface DemoMetrics {
  segment: string;
  subscribers: number;
  marketSharePct: number;
  averageRevenue: number;
  growthRatePct: number;
}

export interface CompetitorMetrics {
  name: string;
  marketSharePct: number;
  subscribers: number;
  averagePrice: number;
  strengths: string[];
  weaknesses: string[];
}

export interface MarketOpportunity {
  id: string;
  type: string;
  description: string;
  potentialSize: number;
  confidence: number;
  requiredActions: string[];
}

// Predictive Maintenance Types
export interface PredictiveMaintenanceMetrics {
  period: string;
  totalAssets: number;
  healthyAssets: number;
  atRiskAssets: number;
  criticalAssets: number;
  overallHealthScore: number;
  byAssetType: Record<string, AssetTypeMetrics>;
  predictedFailures: PredictedFailure[];
  maintenanceSchedule: MaintenanceTask[];
  generatedAt: string;
}

export interface AssetTypeMetrics {
  assetType: string;
  total: number;
  healthy: number;
  atRisk: number;
  critical: number;
  healthScore: number;
}

export interface PredictedFailure {
  assetId: string;
  assetType: string;
  failureType: string;
  predictedDate: string;
  confidence: number;
  recommendedActions: string[];
}

export interface MaintenanceTask {
  id: string;
  assetId: string;
  taskType: string;
  priority: string;
  scheduledDate: string;
  estimatedDurationMinutes: number;
  description: string;
  status: string;
}

// Pricing Optimization Types
export interface PricingOptimizationResult {
  ratePlanId: string;
  currentPrice: number;
  optimalPrice: number;
  strategy: string;
  expectedRevenue: number;
  expectedDemand: number;
  priceChangePct: number;
  reasoning: string[];
  risks: string[];
  recommendations: string[];
  confidence: number;
  generatedAt: string;
}

export interface PricingMetrics {
  period: string;
  totalRatePlans: number;
  optimizedRatePlans: number;
  averagePriceChangePct: number;
  expectedRevenueImpactPct: number;
  churnRateReductionPct: number;
  priceElasticity: number;
  competitiveIndex: number;
  optimizationRoiPct: number;
  generatedAt: string;
}

export interface UpdateSubscriberRequest {
  firstName?: string;
  lastName?: string;
  email?: string;
  organizationId?: string;
  planId?: number;
}

export interface SubscriberAccount {
  imsi: string;
  balance: number;
  dataLimit: number;
  dataUsed: number;
  voiceLimit: number;
  voiceUsed: number;
  smsLimit: number;
  smsUsed: number;
  status: SubscriberStatus;
  lastUpdated: string;
}

export interface UsageEvent {
  imsi: string;
  sessionId: string;
  usageType: UsageType;
  volume: number;
  timestamp: string;
  rate: number;
  cost: number;
}

export enum UsageType {
  Data = 'data',
  Voice = 'voice',
  SMS = 'sms'
}

export interface RatingPlan {
  planId: string;
  name: string;
  dataRate: number;
  voiceRate: number;
  smsRate: number;
  monthlyFee: number;
  dataLimit: number;
  voiceLimit: number;
  smsLimit: number;
}

export interface ChargingSession {
  sessionId: string;
  imsi: string;
  startTime: string;
  endTime?: string;
  dataBytes: number;
  voiceSeconds: number;
  smsCount: number;
  totalCost: number;
  status: SessionStatus;
}

export enum SessionStatus {
  Active = 'active',
  Completed = 'completed',
  Terminated = 'terminated'
}

export interface SystemStats {
  activeSessions: number;
  totalAccounts: number;
  blockedUsers: number;
  lowBalanceAlerts: number;
  uptime: number;
}

export interface HealthStatus {
  redisConnected: boolean;
  activeSync: boolean;
  lastSync: string;
  memoryUsage: number;
}

export interface ApiResponse<T> {
  data: T;
  message: string;
  success: boolean;
  code: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface ListSubscribersRequest {
  page?: number;
  pageSize?: number;
  status?: SubscriberStatus;
  organizationId?: string;
  search?: string;
}

export interface TopUpRequest {
  amount: number;
  paymentMethodId?: string;
}

export interface PaymentMethod {
  id: string;
  type: PaymentMethodType;
  last4: string;
  expiryMonth: number;
  expiryYear: number;
  brand: string;
  isDefault: boolean;
}

export enum PaymentMethodType {
  CreditCard = 'credit_card',
  BankAccount = 'bank_account'
}

export interface Invoice {
  id: string;
  subscriberId: number;
  amount: number;
  currency: string;
  status: InvoiceStatus;
  dueDate: string;
  createdAt: string;
  lineItems: InvoiceLineItem[];
}

export enum InvoiceStatus {
  Draft = 'draft',
  Pending = 'pending',
  Paid = 'paid',
  Overdue = 'overdue',
  Cancelled = 'cancelled'
}

export interface InvoiceLineItem {
  description: string;
  quantity: number;
  unitPrice: number;
  amount: number;
}

export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: string;
}

export interface RealTimeUsageUpdate {
  imsi: string;
  dataUsed: number;
  voiceUsed: number;
  smsUsed: number;
  cost: number;
  timestamp: string;
}

export interface Alert {
  id: string;
  type: AlertType;
  severity: AlertSeverity;
  message: string;
  subscriberId?: number;
  timestamp: string;
  resolved: boolean;
}

export enum AlertType {
  LowBalance = 'low_balance',
  HighUsage = 'high_usage',
  PaymentFailed = 'payment_failed',
  SystemError = 'system_error'
}

export enum AlertSeverity {
  Low = 'low',
  Medium = 'medium',
  High = 'high',
  Critical = 'critical'
}

export interface Error {
  code: string;
  message: string;
  details?: any;
  timestamp: string;
}
