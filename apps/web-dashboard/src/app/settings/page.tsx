import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Settings } from "lucide-react";

export default function SettingsPage() {
  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center gap-3">
        <Settings className="size-8 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
          <p className="text-muted-foreground mt-1">Platform configuration and preferences.</p>
        </div>
      </div>

      <div className="grid gap-6 max-w-2xl">
        <Card>
          <CardHeader>
            <CardTitle>API Configuration</CardTitle>
            <CardDescription>Configure the API server connection.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="text-sm font-medium">API Base URL</label>
              <input type="text" defaultValue="http://localhost:8000" className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" />
            </div>
            <div>
              <label className="text-sm font-medium">Charging Engine URL</label>
              <input type="text" defaultValue="http://localhost:3001" className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" />
            </div>
            <Button size="sm">Save Changes</Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>IMSI Configuration</CardTitle>
            <CardDescription>Mobile Country Code and Mobile Network Code settings.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium">MCC+MNC Prefix</label>
                <input type="text" defaultValue="20893" className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" />
              </div>
              <div>
                <label className="text-sm font-medium">IMSI Range</label>
                <input type="text" defaultValue="1 - 999999999" className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" readOnly />
              </div>
            </div>
            <Button size="sm">Save Changes</Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>ES2+ SM-DP+ Configuration</CardTitle>
            <CardDescription>eSIM profile server connection settings.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="text-sm font-medium">SM-DP+ URL</label>
              <input type="text" defaultValue="https://smdp.example.com" className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" />
            </div>
            <div>
              <label className="text-sm font-medium">API Key</label>
              <input type="password" defaultValue="••••••••" className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" />
            </div>
            <Button size="sm">Save Changes</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
