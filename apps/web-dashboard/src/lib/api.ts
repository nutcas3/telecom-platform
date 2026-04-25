const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";
const CACHE_TTL = parseInt(process.env.NEXT_PUBLIC_CACHE_TTL || "30000");

let authToken: string | null = null;

import { getCachedData, invalidateCachePattern } from "./cache";

export function setAuthToken(token: string) {
  authToken = token;
  if (typeof window !== "undefined") {
    localStorage.setItem("auth_token", token);
  }
}

export function getAuthToken(): string | null {
  if (authToken) return authToken;
  if (typeof window !== "undefined") {
    return localStorage.getItem("auth_token");
  }
  return null;
}

export function clearAuthToken() {
  authToken = null;
  if (typeof window !== "undefined") {
    localStorage.removeItem("auth_token");
  }
  invalidateCachePattern(".*");
}

async function apiFetch<T>(path: string, init?: RequestInit, useCache: boolean = false): Promise<T> {
  if (useCache && init?.method === "GET") {
    try {
      return await getCachedData(path, () => fetchUncached<T>(path, init), CACHE_TTL);
    } catch (e) {
      console.error("Cache fetch failed, falling back to direct fetch:", e);
      return fetchUncached<T>(path, init);
    }
  }
  return fetchUncached<T>(path, init);
}

async function fetchUncached<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getAuthToken();
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { ...headers, ...init?.headers },
    ...init,
  });
  if (!res.ok) {
    if (res.status === 401) {
      clearAuthToken();
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
    }
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || res.statusText);
  }
  return res.json();
}

export interface Subscriber {
  id: number;
  imsi: string;
  msisdn: string;
  first_name: string;
  last_name: string;
  email: string;
  organization_id: string;
  status: "active" | "inactive" | "suspended" | "terminated" | "provisioning";
  plan_id: number;
  balance: number;
  profile_status: string;
  created_at: string;
  updated_at: string;
}

export interface SystemStats {
  active_sessions: number;
  total_accounts: number;
  blocked_users: number;
  low_balance_alerts: number;
  uptime: number;
}

export interface HealthStatus {
  redis_connected: boolean;
  active_sync: boolean;
  last_sync: string;
  memory_usage: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
}

// Services Management Types
export interface Service {
  name: string;
  namespace: string;
  creationTimestamp: string;
  labels: Record<string, string>;
  replicas: number;
  readyReplicas: number;
  availableReplicas: number;
  updatedReplicas: number;
  unavailableReplicas: number;
}

export interface ServiceHealth {
  service: string;
  status: "healthy" | "degraded" | "unhealthy";
  replicas_desired: number;
  replicas_ready: number;
  replicas_available: number;
}

export interface PodStatus {
  deployment: {
    replicas: number;
    ready_replicas: number;
    available_replicas: number;
    updated_replicas: number;
    unavailable_replicas: number;
  };
  pods: Array<{
    name: string;
    phase: string;
    node_name: string;
    created: string;
    conditions: Array<{
      type: string;
      status: string;
      reason: string;
    }>;
  }>;
}

export interface ServiceEvent {
  type: string;
  reason: string;
  message: string;
  first_timestamp: string;
  last_timestamp: string;
  count: number;
}

// Monitoring & Metrics Types
export interface MetricSample {
  timestamp: string;
  value: number;
}

export interface Alert {
  name: string;
  state: string;
  severity: string;
  summary: string;
  description: string;
  startsAt: string;
  endsAt?: string;
}

// Deployment Management Types
export interface Deployment {
  id: number;
  service: string;
  version: string;
  status: "pending" | "running" | "completed" | "failed" | "rolling_back";
  started_at: string;
  completed_at?: string;
  metadata: Record<string, any>;
}

export interface DeploymentStartRequest {
  service: string;
  version: string;
  metadata?: Record<string, any>;
}

// Plugin Management Types
export interface Plugin {
  id: number;
  name: string;
  version: string;
  description: string;
  author: string;
  enabled: boolean;
  config: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface PluginInstallRequest {
  name: string;
  version: string;
  config?: Record<string, any>;
}

// Automation Management Types
export interface Automation {
  id: number;
  name: string;
  description: string;
  trigger: string;
  schedule?: string;
  enabled: boolean;
  config: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface AutomationRun {
  id: number;
  automation_id: number;
  status: "pending" | "running" | "completed" | "failed";
  started_at: string;
  completed_at?: string;
  output?: string;
  error?: string;
}

// Billing & Invoicing Types
export interface Invoice {
  id: number;
  subscriber_id: number;
  amount: number;
  currency: string;
  status: "draft" | "sent" | "paid" | "overdue" | "cancelled";
  due_date: string;
  created_at: string;
  line_items: Array<{
    description: string;
    quantity: number;
    unit_price: number;
    amount: number;
  }>;
}

export interface Payment {
  id: number;
  invoice_id: number;
  amount: number;
  currency: string;
  status: "pending" | "completed" | "failed" | "refunded";
  method: string;
  transaction_id: string;
  created_at: string;
}

// Configuration Management Types
export interface ConfigEntry {
  id: number;
  key: string;
  value: string;
  type: string;
  sensitive: boolean;
  description: string;
  updated_at: string;
}

export interface ConfigValidationResult {
  valid: boolean;
  errors: string[];
}

// Chaos Engineering Types
export interface ChaosExperiment {
  id: number;
  name: string;
  type: string;
  target: string;
  parameters: Record<string, any>;
  status: "pending" | "running" | "completed" | "failed" | "cancelled";
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface ChaosExperimentRequest {
  name: string;
  type: string;
  target: string;
  parameters: Record<string, any>;
}

// Authentication Types
export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  first_name: string;
  last_name: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface User {
  id: number;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export const api = {
  // Authentication
  auth: {
    login: (data: LoginRequest) => 
      apiFetch<AuthResponse>("/v1/auth/login", { method: "POST", body: JSON.stringify(data) }),
    register: (data: RegisterRequest) => 
      apiFetch<AuthResponse>("/v1/auth/register", { method: "POST", body: JSON.stringify(data) }),
    refresh: (refreshToken: string) => 
      apiFetch<AuthResponse>("/v1/auth/refresh", { 
        method: "POST", 
        body: JSON.stringify({ refresh_token: refreshToken }),
        headers: { "Content-Type": "application/json" }
      }),
    logout: () => 
      apiFetch<void>("/v1/auth/logout", { method: "POST" }),
    profile: () => 
      apiFetch<User>("/v1/auth/profile"),
    changePassword: (data: ChangePasswordRequest) => 
      apiFetch<void>("/v1/auth/change-password", { method: "POST", body: JSON.stringify(data) }),
  },
  subscribers: {
    list: (page = 1, pageSize = 20) =>
      apiFetch<PaginatedResponse<Subscriber>>(`/api/subscribers?page=${page}&pageSize=${pageSize}`, undefined, true),
    get: (id: number) => apiFetch<Subscriber>(`/api/subscribers/${id}`, undefined, true),
    create: (data: Partial<Subscriber>) =>
      apiFetch<Subscriber>("/api/subscribers", { method: "POST", body: JSON.stringify(data) }),
    update: (id: number, data: Partial<Subscriber>) =>
      apiFetch<Subscriber>(`/api/subscribers/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    delete: (id: number) =>
      apiFetch<void>(`/api/subscribers/${id}`, { method: "DELETE" }),
  },
  system: {
    stats: () => apiFetch<SystemStats>("/api/system/stats", undefined, true),
    health: () => apiFetch<HealthStatus>("/api/system/health", undefined, true),
  },
  // Services Management
  services: {
    list: () => apiFetch<{ services: Service[]; total: number; namespace: string }>("/v1/services", undefined, true),
    get: (id: string) => apiFetch<Service>(`/v1/services/${id}`, undefined, true),
    restart: (id: string) => apiFetch<{ message: string; service: string }>(`/v1/services/${id}/restart`, { method: "POST" }),
    logs: (id: string, tailLines = 100) => 
      apiFetch<string>(`/v1/services/${id}/logs?tailLines=${tailLines}`),
    start: (id: string) => apiFetch<{ message: string; service: string }>(`/v1/services/${id}/start`, { method: "POST" }),
    stop: (id: string) => apiFetch<{ message: string; service: string }>(`/v1/services/${id}/stop`, { method: "POST" }),
    scale: (id: string, replicas: number) => 
      apiFetch<{ message: string; service: string; replicas: number }>(`/v1/services/${id}/scale`, { 
        method: "POST", 
        body: JSON.stringify({ replicas }) 
      }),
    health: (id: string) => apiFetch<ServiceHealth>(`/v1/services/${id}/health`, undefined, true),
    podStatus: (id: string) => apiFetch<PodStatus>(`/v1/services/${id}/pods`, undefined, true),
    events: (id: string, limit = 50) => 
      apiFetch<{ events: ServiceEvent[]; total: number }>(`/v1/services/${id}/events?limit=${limit}`, undefined, true),
  },
  // Monitoring & Metrics
  monitoring: {
    metrics: (query = "up") => apiFetch<{ data: MetricSample[] }>(`/v1/monitoring/metrics?query=${query}`, undefined, true),
    alerts: () => apiFetch<Alert[]>("/v1/monitoring/alerts", undefined, true),
    health: () => apiFetch<{ status: string; services: Record<string, string> }>("/v1/monitoring/health", undefined, true),
    logs: (service?: string, level?: string, limit = 100) => 
      apiFetch<{ logs: string[]; total: number }>(`/v1/monitoring/logs?service=${service || ""}&level=${level || ""}&limit=${limit}`),
  },
  // Deployment Management
  deployments: {
    status: () => apiFetch<{ deployments: Deployment[]; total: number }>("/v1/deploy/status", undefined, true),
    start: (data: DeploymentStartRequest) => 
      apiFetch<Deployment>("/v1/deploy/start", { method: "POST", body: JSON.stringify(data) }),
    rollback: (deploymentId: number) => 
      apiFetch<Deployment>("/v1/deploy/rollback", { method: "POST", body: JSON.stringify({ deployment_id: deploymentId }) }),
    history: (page = 1, pageSize = 20) => 
      apiFetch<PaginatedResponse<Deployment>>(`/v1/deploy/history?page=${page}&pageSize=${pageSize}`, undefined, true),
  },
  // Plugin Management
  plugins: {
    list: (enabled?: boolean) => 
      apiFetch<{ plugins: Plugin[]; total: number }>(`/v1/plugins${enabled !== undefined ? `?enabled=${enabled}` : ""}`, undefined, true),
    install: (data: PluginInstallRequest) => 
      apiFetch<Plugin>("/v1/plugins/install", { method: "POST", body: JSON.stringify(data) }),
    uninstall: (id: number) => apiFetch<void>(`/v1/plugins/${id}`, { method: "DELETE" }),
    enable: (id: number) => apiFetch<Plugin>(`/v1/plugins/${id}/enable`, { method: "POST" }),
    disable: (id: number) => apiFetch<Plugin>(`/v1/plugins/${id}/disable`, { method: "POST" }),
    get: (id: number) => apiFetch<Plugin>(`/v1/plugins/${id}`, undefined, true),
  },
  // Automation Management
  automation: {
    list: (enabled?: boolean) => 
      apiFetch<{ automations: Automation[]; total: number }>(`/v1/automation${enabled !== undefined ? `?enabled=${enabled}` : ""}`, undefined, true),
    run: (automationId: number, parameters?: Record<string, any>) => 
      apiFetch<AutomationRun>("/v1/automation/run", { method: "POST", body: JSON.stringify({ automation_id: automationId, parameters }) }),
    schedule: (automationId: number, schedule: string) => 
      apiFetch<Automation>(`/v1/automation/${automationId}/schedule`, { method: "POST", body: JSON.stringify({ schedule }) }),
    logs: (automationId?: number, page = 1, pageSize = 20) => 
      apiFetch<PaginatedResponse<AutomationRun>>(`/v1/automation/logs${automationId ? `?automation_id=${automationId}&` : "?"}page=${page}&pageSize=${pageSize}`, undefined, true),
    get: (id: number) => apiFetch<Automation>(`/v1/automation/${id}`, undefined, true),
    create: (data: Partial<Automation>) => 
      apiFetch<Automation>("/v1/automation", { method: "POST", body: JSON.stringify(data) }),
    update: (id: number, data: Partial<Automation>) => 
      apiFetch<Automation>(`/v1/automation/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    delete: (id: number) => apiFetch<void>(`/v1/automation/${id}`, { method: "DELETE" }),
  },
  // Billing & Invoicing
  billing: {
    invoices: (subscriberId?: number, status?: string, page = 1, pageSize = 20) => 
      apiFetch<PaginatedResponse<Invoice>>(`/v1/billing/invoices${subscriberId ? `?subscriber_id=${subscriberId}&` : "?"}${status ? `status=${status}&` : ""}page=${page}&pageSize=${pageSize}`, undefined, true),
    payments: (invoiceId?: number, status?: string, page = 1, pageSize = 20) => 
      apiFetch<PaginatedResponse<Payment>>(`/v1/billing/payments${invoiceId ? `?invoice_id=${invoiceId}&` : "?"}${status ? `status=${status}&` : ""}page=${page}&pageSize=${pageSize}`, undefined, true),
    generateInvoice: (subscriberId: number, month: string, year: number) => 
      apiFetch<Invoice>("/v1/billing/invoices", { method: "POST", body: JSON.stringify({ subscriber_id: subscriberId, month, year }) }),
    getInvoice: (id: number) => apiFetch<Invoice>(`/v1/billing/invoices/${id}`, undefined, true),
  },
  // Configuration Management
  config: {
    get: (key?: string) => apiFetch<ConfigEntry[] | ConfigEntry>(key ? `/v1/config?key=${key}` : "/v1/config", undefined, true),
    update: (key: string, value: string, type = "string") => 
      apiFetch<ConfigEntry>("/v1/config", { method: "POST", body: JSON.stringify({ key, value, type }) }),
    validate: (config: Record<string, any>) => 
      apiFetch<ConfigValidationResult>("/v1/config/validate", { method: "POST", body: JSON.stringify(config) }),
    delete: (key: string) => apiFetch<void>(`/v1/config/${key}`, { method: "DELETE" }),
  },
  // Chaos Engineering
  chaos: {
    experiments: (status?: string, page = 1, pageSize = 20) => 
      apiFetch<PaginatedResponse<ChaosExperiment>>(`/v1/chaos/experiments${status ? `?status=${status}&` : "?"}page=${page}&pageSize=${pageSize}`, undefined, true),
    run: (data: ChaosExperimentRequest) => 
      apiFetch<ChaosExperiment>("/v1/chaos/experiments", { method: "POST", body: JSON.stringify(data) }),
    status: (experimentId: number) => apiFetch<ChaosExperiment>(`/v1/chaos/status?experiment_id=${experimentId}`, undefined, true),
    cancel: (experimentId: number) => 
      apiFetch<ChaosExperiment>(`/v1/chaos/experiments/${experimentId}`, { method: "DELETE" }),
  },
};
