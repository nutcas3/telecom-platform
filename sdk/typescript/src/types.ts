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
