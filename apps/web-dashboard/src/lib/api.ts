const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...init?.headers },
    ...init,
  });
  if (!res.ok) {
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

export const api = {
  subscribers: {
    list: (page = 1, pageSize = 20) =>
      apiFetch<PaginatedResponse<Subscriber>>(`/api/subscribers?page=${page}&pageSize=${pageSize}`),
    get: (id: number) => apiFetch<Subscriber>(`/api/subscribers/${id}`),
    create: (data: Partial<Subscriber>) =>
      apiFetch<Subscriber>("/api/subscribers", { method: "POST", body: JSON.stringify(data) }),
    update: (id: number, data: Partial<Subscriber>) =>
      apiFetch<Subscriber>(`/api/subscribers/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    delete: (id: number) =>
      apiFetch<void>(`/api/subscribers/${id}`, { method: "DELETE" }),
  },
  system: {
    stats: () => apiFetch<SystemStats>("/api/system/stats"),
    health: () => apiFetch<HealthStatus>("/api/system/health"),
  },
};
