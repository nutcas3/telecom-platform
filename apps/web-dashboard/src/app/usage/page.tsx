import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatCard } from "@/components/stat-card";
import { Activity, HardDrive, Phone, MessageSquare } from "lucide-react";

const usageEvents = [
  { id: 1, imsi: "208930000000001", type: "DATA", volume: "1.2 GB", cost: "€2.40", time: "2 min ago" },
  { id: 2, imsi: "208930000000002", type: "VOICE", volume: "15 min", cost: "€0.75", time: "5 min ago" },
  { id: 3, imsi: "208930000000005", type: "SMS", volume: "3 msgs", cost: "€0.30", time: "12 min ago" },
  { id: 4, imsi: "208930000000001", type: "DATA", volume: "450 MB", cost: "€0.90", time: "18 min ago" },
  { id: 5, imsi: "208930000000008", type: "DATA", volume: "2.8 GB", cost: "€5.60", time: "25 min ago" },
  { id: 6, imsi: "208930000000006", type: "VOICE", volume: "42 min", cost: "€2.10", time: "30 min ago" },
  { id: 7, imsi: "208930000000002", type: "DATA", volume: "780 MB", cost: "€1.56", time: "45 min ago" },
  { id: 8, imsi: "208930000000005", type: "SMS", volume: "1 msg", cost: "€0.10", time: "1 hr ago" },
];

const typeIcon = (t: string) => {
  if (t === "DATA") return HardDrive;
  if (t === "VOICE") return Phone;
  return MessageSquare;
};

export default function UsagePage() {
  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Usage &amp; Billing</h1>
        <p className="text-muted-foreground mt-1">Monitor real-time usage and billing metrics.</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Data Usage Today" value="48.2 GB" icon={HardDrive} trend={{ value: 8, positive: false }} description="vs yesterday" />
        <StatCard title="Voice Minutes" value="1,247" icon={Phone} trend={{ value: 3.1, positive: true }} description="vs yesterday" />
        <StatCard title="SMS Sent" value="389" icon={MessageSquare} description="today" />
        <StatCard title="Revenue Today" value="€1,842" icon={Activity} trend={{ value: 15, positive: true }} description="vs yesterday" />
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
