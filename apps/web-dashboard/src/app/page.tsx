"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatCard } from "@/components/stat-card";
import { Users, Activity, AlertTriangle, Server, Wifi, Database } from "lucide-react";
import { api, SystemStats, HealthStatus, Subscriber } from "@/lib/api";

export default function DashboardPage() {
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [subscribers, setSubscribers] = useState<Subscriber[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadData = async () => {
      try {
        setLoading(true);
        const [statsData, healthData, subscribersData] = await Promise.all([
          api.system.stats(),
          api.system.health(),
          api.subscribers.list(1, 5)
        ]);
        
        setStats(statsData);
        setHealth(healthData);
        setSubscribers(subscribersData.data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load dashboard data");
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, []);

  if (loading) {
    return (
      <div className="p-8 space-y-8">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-48 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-96"></div>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="animate-pulse">
              <div className="h-24 bg-gray-200 rounded"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h3 className="text-red-800 font-medium">Error loading dashboard</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground mt-1">Telecom platform overview and real-time metrics.</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Subscribers"
          value={stats?.total_accounts?.toLocaleString() || "0"}
          icon={Users}
          description="total accounts"
        />
        <StatCard
          title="Active Sessions"
          value={stats?.active_sessions?.toLocaleString() || "0"}
          icon={Activity}
          description="currently active"
        />
        <StatCard
          title="Low Balance Alerts"
          value={stats?.low_balance_alerts?.toLocaleString() || "0"}
          icon={AlertTriangle}
          description="need attention"
        />
        <StatCard
          title="System Uptime"
          value={`${((stats?.uptime || 0) / (24 * 3600 * 1000)).toFixed(1)} days`}
          icon={Server}
          description="continuous operation"
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Recent Subscribers</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {subscribers.length > 0 ? (
                subscribers.map((sub) => (
                  <div key={sub.id} className="flex items-center justify-between">
                    <div>
                      <p className="font-medium text-sm">{sub.first_name} {sub.last_name}</p>
                      <p className="text-xs text-muted-foreground">{sub.msisdn}</p>
                    </div>
                    <Badge
                      variant={
                        sub.status === "active" ? "success"
                          : sub.status === "suspended" ? "warning"
                          : sub.status === "terminated" ? "destructive"
                          : "secondary"
                      }
                    >
                      {sub.status}
                    </Badge>
                  </div>
                ))
              ) : (
                <p className="text-sm text-muted-foreground">No recent subscribers found</p>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {[
                { label: "API Server", icon: Server, status: health ? "Connected" : "Disconnected", ok: true },
                { label: "Database", icon: Database, status: "Connected", ok: true },
                { label: "Redis Cache", icon: Database, status: health?.redis_connected ? "Connected" : "Disconnected", ok: health?.redis_connected },
                { label: "Charging Engine", icon: Activity, status: health?.active_sync ? "Running" : "Stopped", ok: health?.active_sync },
                { label: "Memory Usage", icon: Server, status: `${health?.memory_usage || 0}%`, ok: (health?.memory_usage || 0) < 80 },
              ].map((svc) => (
                <div key={svc.label} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <svc.icon className="size-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{svc.label}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`size-2 rounded-full ${svc.ok ? "bg-emerald-500" : "bg-red-500"}`} />
                    <span className="text-xs text-muted-foreground">{svc.status}</span>
                  </div>
                </div>
              ))}
              {health?.last_sync && (
                <div className="text-xs text-muted-foreground pt-2 border-t">
                  Last sync: {new Date(health.last_sync).toLocaleString()}
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
