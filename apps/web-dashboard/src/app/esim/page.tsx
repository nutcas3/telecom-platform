"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Radio, Download, RefreshCw } from "lucide-react";
import { api, Subscriber } from "@/lib/api";

interface ESIMProfile {
  id: string;
  imsi: string;
  euiccId: string;
  status: "active" | "downloading" | "inactive" | "failed";
  subscriber: string;
  activatedAt: string;
}

const statusVariant = (s: string) =>
  s === "active" ? "success" : s === "downloading" ? "secondary" : s === "failed" ? "destructive" : "warning";

export default function ESIMPage() {
  const [profiles, setProfiles] = useState<ESIMProfile[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadProfiles = async () => {
      try {
        setLoading(true);
        const subscribersResponse = await api.subscribers.list(1, 100);
        
        // Map subscribers to eSIM profiles (mock data structure for now)
        const esimProfiles: ESIMProfile[] = subscribersResponse.data.map((subscriber) => ({
          id: `prof_${subscriber.id.toString().padStart(3, '0')}`,
          imsi: subscriber.imsi,
          euiccId: `890490320040088826${subscriber.id.toString().padStart(5, '0')}`,
          status: subscriber.profile_status === "active" ? "active" : 
                 subscriber.status === "provisioning" ? "downloading" : 
                 "inactive",
          subscriber: `${subscriber.first_name} ${subscriber.last_name}`,
          activatedAt: subscriber.created_at.split('T')[0]
        }));

        setProfiles(esimProfiles);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load eSIM profiles");
      } finally {
        setLoading(false);
      }
    };

    loadProfiles();
  }, []);

  if (loading) {
    return (
      <div className="p-8 space-y-6">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-48 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-96"></div>
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
          <h3 className="text-red-800 font-medium">Error loading eSIM profiles</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Radio className="size-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">eSIM Profiles</h1>
            <p className="text-muted-foreground mt-1">Manage eSIM profile provisioning via SM-DP+ (ES2+).</p>
          </div>
        </div>
        <Button size="sm"><Download className="size-4 mr-1.5" />Provision New</Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Profiles</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-3 font-medium">Profile ID</th>
                  <th className="pb-3 font-medium">Subscriber</th>
                  <th className="pb-3 font-medium">IMSI</th>
                  <th className="pb-3 font-medium">eUICC ID</th>
                  <th className="pb-3 font-medium">Status</th>
                  <th className="pb-3 font-medium text-right">Activated</th>
                  <th className="pb-3 font-medium text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {profiles.map((p) => (
                  <tr key={p.id} className="border-b last:border-0 hover:bg-muted/50 transition-colors">
                    <td className="py-3 font-mono text-xs">{p.id}</td>
                    <td className="py-3">{p.subscriber}</td>
                    <td className="py-3 font-mono text-xs">{p.imsi}</td>
                    <td className="py-3 font-mono text-xs truncate max-w-[180px]">{p.euiccId}</td>
                    <td className="py-3"><Badge variant={statusVariant(p.status)}>{p.status}</Badge></td>
                    <td className="py-3 text-right text-muted-foreground text-xs">{p.activatedAt}</td>
                    <td className="py-3 text-right">
                      <Button variant="ghost" size="xs"><RefreshCw className="size-3" /></Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
