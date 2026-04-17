import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { UserPlus, Search, Download } from "lucide-react";

const subscribers = [
  { id: 1, name: "Alice Martin", msisdn: "+33612345678", imsi: "208930000000001", email: "alice@example.com", status: "active" as const, balance: 45.0, plan: "Premium 5G" },
  { id: 2, name: "Bob Dupont", msisdn: "+33698765432", imsi: "208930000000002", email: "bob@example.com", status: "active" as const, balance: 12.5, plan: "Basic LTE" },
  { id: 3, name: "Claire Moreau", msisdn: "+33678901234", imsi: "208930000000003", email: "claire@example.com", status: "provisioning" as const, balance: 0.0, plan: "Premium 5G" },
  { id: 4, name: "David Leroy", msisdn: "+33667890123", imsi: "208930000000004", email: "david@example.com", status: "suspended" as const, balance: -2.3, plan: "Standard" },
  { id: 5, name: "Emma Petit", msisdn: "+33656789012", imsi: "208930000000005", email: "emma@example.com", status: "active" as const, balance: 78.9, plan: "Premium 5G" },
  { id: 6, name: "François Blanc", msisdn: "+33645678901", imsi: "208930000000006", email: "francois@example.com", status: "active" as const, balance: 25.0, plan: "Standard" },
  { id: 7, name: "Gabrielle Roux", msisdn: "+33634567890", imsi: "208930000000007", email: "gabrielle@example.com", status: "terminated" as const, balance: 0.0, plan: "Basic LTE" },
  { id: 8, name: "Henri Faure", msisdn: "+33623456789", imsi: "208930000000008", email: "henri@example.com", status: "active" as const, balance: 156.2, plan: "Enterprise" },
];

const statusVariant = (s: string) =>
  s === "active" ? "success" : s === "suspended" ? "warning" : s === "terminated" ? "destructive" : "secondary";

export default function SubscribersPage() {
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
                {subscribers.map((sub) => (
                  <tr key={sub.id} className="border-b last:border-0 hover:bg-muted/50 transition-colors">
                    <td className="py-3">
                      <div>
                        <p className="font-medium">{sub.name}</p>
                        <p className="text-xs text-muted-foreground">{sub.email}</p>
                      </div>
                    </td>
                    <td className="py-3 font-mono text-xs">{sub.msisdn}</td>
                    <td className="py-3 font-mono text-xs">{sub.imsi}</td>
                    <td className="py-3">{sub.plan}</td>
                    <td className="py-3 text-right font-mono">
                      <span className={sub.balance < 0 ? "text-red-600" : ""}>
                        €{sub.balance.toFixed(2)}
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
