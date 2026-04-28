"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatCard } from "@/components/stat-card";
import { Activity, HardDrive, Phone, MessageSquare } from "lucide-react";
import { api } from "@/lib/api";

interface UsageEvent {
  id: number;
  imsi: string;
  type: "DATA" | "VOICE" | "SMS";
  volume: string;
  cost: string;
  time: string;
}

interface UsageStats {
  data_usage_today: string;
  voice_minutes_today: number;
  sms_sent_today: number;
  revenue_today: string;
}

const typeIcon = (t: string) => {
  if (t === "DATA") return HardDrive;
  if (t === "VOICE") return Phone;
  return MessageSquare;
};

export default function UsagePage() {
  const [usageEvents, setUsageEvents] = useState<UsageEvent[]>([]);
  const [stats, setStats] = useState<UsageStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadUsageData = async () => {
      try {
        setLoading(true);
        
        // Get monitoring logs for usage events
        const logsResponse = await api.monitoring.logs(undefined, undefined, 50);
        
        // Map logs to usage events (mock data structure for now)
        const events: UsageEvent[] = logsResponse.logs.map((log, index) => {
          const parts = log.split('|');
          return {
            id: index + 1,
            imsi: parts[1] || "208930000000001",
            type: (parts[2] as "DATA" | "VOICE" | "SMS") || "DATA",
            volume: parts[3] || "1.2 GB",
            cost: parts[4] || "EUR2.40",
            time: parts[5] || "2 min ago"
          };
        });

        // Mock stats for now (would come from real metrics endpoint)
        const mockStats: UsageStats = {
          data_usage_today: "48.2 GB",
          voice_minutes_today: 1247,
          sms_sent_today: 389,
          revenue_today: "EUR1,842"
        };

        setUsageEvents(events);
        setStats(mockStats);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load usage data");
      } finally {
        setLoading(false);
      }
    };

    loadUsageData();
  }, []);

  if (loading) {
    return (
      <div className="p-8 space-y-6">
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
        <div className="animate-pulse">
          <div className="h-96 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h3 className="text-red-800 font-medium">Error loading usage data</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Usage &amp; Billing</h1>
        <p className="text-muted-foreground mt-1">Monitor real-time usage and billing metrics.</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Data Usage Today" value={stats?.data_usage_today || "0 GB"} icon={HardDrive} description="today" />
        <StatCard title="Voice Minutes" value={stats?.voice_minutes_today?.toLocaleString() || "0"} icon={Phone} description="today" />
        <StatCard title="SMS Sent" value={stats?.sms_sent_today?.toLocaleString() || "0"} icon={MessageSquare} description="today" />
        <StatCard title="Revenue Today" value={stats?.revenue_today || "EUR0"} icon={Activity} description="today" />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Usage Events</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-3 font-medium">Type</th>
                  <th className="pb-3 font-medium">IMSI</th>
                  <th className="pb-3 font-medium">Volume</th>
                  <th className="pb-3 font-medium text-right">Cost</th>
                  <th className="pb-3 font-medium text-right">Time</th>
                </tr>
              </thead>
              <tbody>
                {usageEvents.map((evt) => {
                  const Icon = typeIcon(evt.type);
                  return (
                    <tr key={evt.id} className="border-b last:border-0 hover:bg-muted/50 transition-colors">
                      <td className="py-3">
                        <div className="flex items-center gap-2">
                          <Icon className="size-4 text-muted-foreground" />
                          <span>{evt.type}</span>
                        </div>
                      </td>
                      <td className="py-3 font-mono text-xs">{evt.imsi}</td>
                      <td className="py-3">{evt.volume}</td>
                      <td className="py-3 text-right font-mono">{evt.cost}</td>
                      <td className="py-3 text-right text-muted-foreground">{evt.time}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
