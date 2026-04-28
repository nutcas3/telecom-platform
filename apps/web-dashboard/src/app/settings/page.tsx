"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Settings } from "lucide-react";
import { api, ConfigEntry } from "@/lib/api";

export default function SettingsPage() {
  const [configs, setConfigs] = useState<ConfigEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const loadConfigs = async () => {
      try {
        setLoading(true);
        const configsData = await api.config.get() as ConfigEntry[];
        setConfigs(configsData);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load configuration");
      } finally {
        setLoading(false);
      }
    };

    loadConfigs();
  }, []);

  const handleSaveConfig = async (key: string, value: string) => {
    try {
      setSaving(true);
      await api.config.update(key, value);
      // Reload configs
      const configsData = await api.config.get() as ConfigEntry[];
      setConfigs(configsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save configuration");
    } finally {
      setSaving(false);
    }
  };

  const getConfigValue = (key: string) => {
    const config = configs.find(c => c.key === key);
    return config?.value || "";
  };

  if (loading) {
    return (
      <div className="p-8 space-y-6">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-48 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-96"></div>
        </div>
        <div className="grid gap-6 max-w-2xl">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="animate-pulse">
              <div className="h-48 bg-gray-200 rounded"></div>
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
          <h3 className="text-red-800 font-medium">Error loading settings</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

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
              <input 
                type="text" 
                defaultValue={getConfigValue("api_base_url") || "http://localhost:8000"} 
                className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" 
              />
            </div>
            <div>
              <label className="text-sm font-medium">Charging Engine URL</label>
              <input 
                type="text" 
                defaultValue={getConfigValue("charging_engine_url") || "http://localhost:3001"} 
                className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" 
              />
            </div>
            <Button size="sm" disabled={saving} onClick={() => {
              const apiUrl = (document.querySelector('input[type="text"]') as HTMLInputElement)?.value || "http://localhost:8000";
              handleSaveConfig("api_base_url", apiUrl);
            }}>
              {saving ? "Saving..." : "Save Changes"}
            </Button>
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
                <input 
                  type="text" 
                  defaultValue={getConfigValue("mcc_mnc_prefix") || "20893"} 
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" 
                />
              </div>
              <div>
                <label className="text-sm font-medium">IMSI Range</label>
                <input 
                  type="text" 
                  defaultValue={getConfigValue("imsi_range") || "1 - 999999999"} 
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" 
                  readOnly 
                />
              </div>
            </div>
            <Button size="sm" disabled={saving} onClick={() => {
              const mccMnc = (document.querySelectorAll('input[type="text"]')[1] as HTMLInputElement)?.value || "20893";
              handleSaveConfig("mcc_mnc_prefix", mccMnc);
            }}>
              {saving ? "Saving..." : "Save Changes"}
            </Button>
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
              <input 
                type="text" 
                defaultValue={getConfigValue("smdp_url") || "https://smdp.example.com"} 
                className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" 
              />
            </div>
            <div>
              <label className="text-sm font-medium">API Key</label>
              <input 
                type="password" 
                defaultValue={getConfigValue("smdp_api_key") ? "" : ""} 
                placeholder={getConfigValue("smdp_api_key") ? "..." : "Enter API key"}
                className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50" 
              />
            </div>
            <Button size="sm" disabled={saving} onClick={() => {
              const smdpUrl = (document.querySelectorAll('input[type="text"]')[2] as HTMLInputElement)?.value || "https://smdp.example.com";
              handleSaveConfig("smdp_url", smdpUrl);
            }}>
              {saving ? "Saving..." : "Save Changes"}
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
