import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { HeartPulse, Server, Database, Cpu, HardDrive, Wifi } from "lucide-react";

const services = [
  { name: "API Server", icon: Server, status: "healthy", latency: "12ms", uptime: "99.99%", cpu: "23%", memory: "512 MB" },
  { name: "PostgreSQL", icon: Database, status: "healthy", latency: "3ms", uptime: "99.97%", cpu: "15%", memory: "2.1 GB" },
  { name: "Redis Cache", icon: Database, status: "healthy", latency: "1ms", uptime: "99.99%", cpu: "5%", memory: "256 MB" },
  { name: "Charging Engine (Rust)", icon: Cpu, status: "healthy", latency: "2ms", uptime: "99.98%", cpu: "8%", memory: "64 MB" },
  { name: "AMF Gateway", icon: Wifi, status: "healthy", latency: "45ms", uptime: "99.95%", cpu: "12%", memory: "128 MB" },
  { name: "ES2+ SM-DP+", icon: HardDrive, status: "degraded", latency: "230ms", uptime: "98.50%", cpu: "N/A", memory: "N/A" },
];

export default function HealthPage() {
  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center gap-3">
        <HeartPulse className="size-8 text-emerald-600" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">System Health</h1>
          <p className="text-muted-foreground mt-1">Monitor service health, latency, and resource usage.</p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {services.map((svc) => (
          <Card key={svc.name}>
            <CardHeader className="flex-row items-center justify-between pb-2">
              <div className="flex items-center gap-2">
                <svc.icon className="size-5 text-muted-foreground" />
                <CardTitle className="text-base">{svc.name}</CardTitle>
              </div>
              <Badge variant={svc.status === "healthy" ? "success" : "warning"}>
                {svc.status}
              </Badge>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-y-2 text-sm">
                <div>
                  <p className="text-muted-foreground text-xs">Latency</p>
                  <p className="font-mono font-medium">{svc.latency}</p>
                </div>
                <div>
                  <p className="text-muted-foreground text-xs">Uptime</p>
                  <p className="font-mono font-medium">{svc.uptime}</p>
                </div>
                <div>
                  <p className="text-muted-foreground text-xs">CPU</p>
                  <p className="font-mono font-medium">{svc.cpu}</p>
                </div>
                <div>
                  <p className="text-muted-foreground text-xs">Memory</p>
                  <p className="font-mono font-medium">{svc.memory}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
