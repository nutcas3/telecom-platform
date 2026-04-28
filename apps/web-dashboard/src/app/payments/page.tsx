"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatCard } from "@/components/stat-card";
import { CreditCard, ArrowUpRight, ArrowDownRight, Clock } from "lucide-react";
import { api, Payment, PaginatedResponse } from "@/lib/api";

interface PaymentStats {
  revenue_today: string;
  refunds_today: string;
  pending_count: number;
  success_rate: string;
}

export default function PaymentsPage() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [stats, setStats] = useState<PaymentStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadPaymentData = async () => {
      try {
        setLoading(true);
        const [paymentsResponse] = await Promise.all([
          api.billing.payments(undefined, undefined, 1, 50)
        ]);

        setPayments(paymentsResponse.data);

        // Calculate stats from payments data
        const today = new Date().toISOString().split('T')[0];
        const todayPayments = paymentsResponse.data.filter(p => 
          p.created_at.startsWith(today)
        );
        
        const revenue = todayPayments
          .filter(p => p.status === 'completed' && p.amount > 0)
          .reduce((sum, p) => sum + p.amount, 0);
        
        const refunds = todayPayments
          .filter(p => p.status === 'completed' && p.amount < 0)
          .reduce((sum, p) => sum + Math.abs(p.amount), 0);
        
        const pending = todayPayments.filter(p => p.status === 'pending').length;
        
        const completed = paymentsResponse.data.filter(p => p.status === 'completed').length;
        const total = paymentsResponse.data.length;
        const successRate = total > 0 ? ((completed / total) * 100).toFixed(1) : '0';

        setStats({
          revenue_today: `EUR${revenue.toFixed(2)}`,
          refunds_today: `EUR${refunds.toFixed(2)}`,
          pending_count: pending,
          success_rate: `${successRate}%`
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load payment data");
      } finally {
        setLoading(false);
      }
    };

    loadPaymentData();
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
          <h3 className="text-red-800 font-medium">Error loading payment data</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Payments</h1>
        <p className="text-muted-foreground mt-1">Transaction history and payment management.</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Revenue Today" value={stats?.revenue_today || "EUR0"} icon={ArrowUpRight} description="today" />
        <StatCard title="Refunds Today" value={stats?.refunds_today || "EUR0"} icon={ArrowDownRight} description="today" />
        <StatCard title="Pending" value={stats?.pending_count?.toString() || "0"} icon={Clock} description="transactions" />
        <StatCard title="Success Rate" value={stats?.success_rate || "0%"} icon={CreditCard} description="all time" />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Transactions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-3 font-medium">Transaction ID</th>
                  <th className="pb-3 font-medium">Subscriber</th>
                  <th className="pb-3 font-medium">Type</th>
                  <th className="pb-3 font-medium text-right">Amount</th>
                  <th className="pb-3 font-medium">Status</th>
                  <th className="pb-3 font-medium text-right">Date</th>
                </tr>
              </thead>
              <tbody>
                {payments.map((payment) => (
                  <tr key={payment.id} className="border-b last:border-0 hover:bg-muted/50 transition-colors">
                    <td className="py-3 font-mono text-xs">{payment.transaction_id}</td>
                    <td className="py-3">Invoice #{payment.invoice_id}</td>
                    <td className="py-3">
                      <Badge variant="secondary">{payment.method}</Badge>
                    </td>
                    <td className="py-3 text-right font-mono">
                      <span className={payment.amount < 0 ? "text-red-600" : ""}>EUR{Math.abs(payment.amount).toFixed(2)}</span>
                    </td>
                    <td className="py-3">
                      <Badge variant={
                        payment.status === "completed" ? "success" : 
                        payment.status === "failed" ? "destructive" : 
                        "secondary"
                      }>
                        {payment.status}
                      </Badge>
                    </td>
                    <td className="py-3 text-right text-muted-foreground text-xs">
                      {new Date(payment.created_at).toLocaleString()}
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
