"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { UserPlus, Search, Download } from "lucide-react";
import { api, Subscriber, PaginatedResponse } from "@/lib/api";

const statusVariant = (s: string) =>
  s === "active" ? "success" : s === "suspended" ? "warning" : s === "terminated" ? "destructive" : "secondary";

export default function SubscribersPage() {
  const [subscribers, setSubscribers] = useState<Subscriber[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    const loadSubscribers = async () => {
      try {
        setLoading(true);
        const response = await api.subscribers.list(currentPage, 20);
        setSubscribers(response.data);
        setTotalPages(Math.ceil(response.total / response.page_size));
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load subscribers");
      } finally {
        setLoading(false);
      }
    };

    loadSubscribers();
  }, [currentPage]);

  const filteredSubscribers = subscribers.filter(sub =>
    sub.first_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    sub.last_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    sub.msisdn.includes(searchTerm) ||
    sub.email.toLowerCase().includes(searchTerm.toLowerCase())
  );

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
          <h3 className="text-red-800 font-medium">Error loading subscribers</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Subscribers</h1>
          <p className="text-muted-foreground mt-1">Manage subscriber accounts and profiles.</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm"><Download className="size-4 mr-1.5" />Export</Button>
          <Button size="sm"><UserPlus className="size-4 mr-1.5" />Add Subscriber</Button>
        </div>
      </div>

      <Card>
        <CardHeader className="flex-row items-center justify-between">
          <CardTitle>All Subscribers</CardTitle>
          <div className="flex items-center gap-2">
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search subscribers..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="h-9 rounded-lg border border-input bg-background pl-9 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring/50 w-64"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-3 font-medium">Name</th>
                  <th className="pb-3 font-medium">MSISDN</th>
                  <th className="pb-3 font-medium">IMSI</th>
                  <th className="pb-3 font-medium">Plan</th>
                  <th className="pb-3 font-medium text-right">Balance</th>
                  <th className="pb-3 font-medium">Status</th>
                  <th className="pb-3 font-medium text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredSubscribers.map((sub) => (
                  <tr key={sub.id} className="border-b last:border-0 hover:bg-muted/50 transition-colors">
                    <td className="py-3">
                      <div>
                        <p className="font-medium">{sub.first_name} {sub.last_name}</p>
                        <p className="text-xs text-muted-foreground">{sub.email}</p>
                      </div>
                    </td>
                    <td className="py-3 font-mono text-xs">{sub.msisdn}</td>
                    <td className="py-3 font-mono text-xs">{sub.imsi}</td>
                    <td className="py-3">Plan {sub.plan_id}</td>
                    <td className="py-3 text-right font-mono">
                      <span className={sub.balance < 0 ? "text-red-600" : ""}>
                        ${sub.balance.toFixed(2)}
                      </span>
                    </td>
                    <td className="py-3">
                      <Badge variant={statusVariant(sub.status)}>{sub.status}</Badge>
                    </td>
                    <td className="py-3 text-right">
                      <Button variant="ghost" size="xs">View</Button>
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
