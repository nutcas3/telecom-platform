"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Puzzle, Download, Power, PowerOff, Settings, Trash2, Search, Plus } from "lucide-react";
import { api, Plugin } from "@/lib/api";

export default function PluginsPage() {
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showInstallForm, setShowInstallForm] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [filterEnabled, setFilterEnabled] = useState<boolean | undefined>(undefined);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [installForm, setInstallForm] = useState({
    name: "",
    version: "",
    config: "{}"
  });

  useEffect(() => {
    loadPlugins();
  }, [filterEnabled]);

  const loadPlugins = async () => {
    try {
      setLoading(true);
      const pluginsResponse = await api.plugins.list(filterEnabled);
      setPlugins(pluginsResponse.plugins);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load plugins");
    } finally {
      setLoading(false);
    }
  };

  const handleInstallPlugin = async () => {
    try {
      setActionLoading('install');
      const config = installForm.config ? JSON.parse(installForm.config) : {};
      await api.plugins.install({
        name: installForm.name,
        version: installForm.version,
        config
      });
      
      // Reset form
      setInstallForm({ name: "", version: "", config: "{}" });
      setShowInstallForm(false);
      await loadPlugins();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to install plugin");
    } finally {
      setActionLoading(null);
    }
  };

  const handleTogglePlugin = async (pluginId: number, enabled: boolean) => {
    try {
      setActionLoading(`toggle-${pluginId}`);
      if (enabled) {
        await api.plugins.enable(pluginId);
      } else {
        await api.plugins.disable(pluginId);
      }
      await loadPlugins();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to toggle plugin");
    } finally {
      setActionLoading(null);
    }
  };

  const handleUninstallPlugin = async (pluginId: number) => {
    if (!confirm("Are you sure you want to uninstall this plugin?")) return;
    
    try {
      setActionLoading(`uninstall-${pluginId}`);
      await api.plugins.uninstall(pluginId);
      await loadPlugins();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to uninstall plugin");
    } finally {
      setActionLoading(null);
    }
  };

  const filteredPlugins = plugins.filter(plugin =>
    plugin.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    plugin.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
    plugin.author.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (loading) {
    return (
      <div className="p-8 space-y-6">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-48 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-96"></div>
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[...Array(6)].map((_, i) => (
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
          <h3 className="text-red-800 font-medium">Error loading plugins</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Puzzle className="size-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Plugins</h1>
            <p className="text-muted-foreground mt-1">Plugin marketplace and management.</p>
          </div>
        </div>
        <Button onClick={() => setShowInstallForm(true)}>
          <Plus className="size-4 mr-2" />
          Install Plugin
        </Button>
      </div>

      {showInstallForm && (
        <Card>
          <CardHeader>
            <CardTitle>Install New Plugin</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium">Plugin Name</label>
                <input
                  type="text"
                  value={installForm.name}
                  onChange={(e) => setInstallForm({ ...installForm, name: e.target.value })}
                  placeholder="e.g., rate-limiter"
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Version</label>
                <input
                  type="text"
                  value={installForm.version}
                  onChange={(e) => setInstallForm({ ...installForm, version: e.target.value })}
                  placeholder="e.g., v1.0.0"
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Configuration (JSON)</label>
                <textarea
                  value={installForm.config}
                  onChange={(e) => setInstallForm({ ...installForm, config: e.target.value })}
                  placeholder='{"rate_limit": 100, "window": "1m"}'
                  className="mt-1 block w-full h-20 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div className="flex gap-2">
                <Button onClick={handleInstallPlugin} disabled={actionLoading !== null || !installForm.name || !installForm.version}>
                  {actionLoading === 'install' ? 'Installing...' : 'Install Plugin'}
                </Button>
                <Button variant="outline" onClick={() => setShowInstallForm(false)}>
                  Cancel
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardContent className="p-4">
          <div className="flex items-center gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search plugins..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="h-9 rounded-lg border border-input bg-background pl-9 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring/50 w-full"
              />
            </div>
            <div className="flex gap-2">
              <Button
                variant={filterEnabled === undefined ? "default" : "outline"}
                size="sm"
                onClick={() => setFilterEnabled(undefined)}
              >
                All
              </Button>
              <Button
                variant={filterEnabled === true ? "default" : "outline"}
                size="sm"
                onClick={() => setFilterEnabled(true)}
              >
                Enabled
              </Button>
              <Button
                variant={filterEnabled === false ? "default" : "outline"}
                size="sm"
                onClick={() => setFilterEnabled(false)}
              >
                Disabled
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Plugin Statistics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <Puzzle className="size-5 text-blue-500" />
              <div>
                <p className="text-2xl font-bold">{plugins.length}</p>
                <p className="text-sm text-muted-foreground">Total Plugins</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <Power className="size-5 text-green-500" />
              <div>
                <p className="text-2xl font-bold">{plugins.filter(p => p.enabled).length}</p>
                <p className="text-sm text-muted-foreground">Enabled</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <PowerOff className="size-5 text-gray-500" />
              <div>
                <p className="text-2xl font-bold">{plugins.filter(p => !p.enabled).length}</p>
                <p className="text-sm text-muted-foreground">Disabled</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <Settings className="size-5 text-purple-500" />
              <div>
                <p className="text-2xl font-bold">{plugins.filter(p => Object.keys(p.config).length > 0).length}</p>
                <p className="text-sm text-muted-foreground">Configured</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Plugins Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {filteredPlugins.map((plugin) => (
          <Card key={plugin.id} className="relative">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">{plugin.name}</CardTitle>
                <Badge variant={plugin.enabled ? "success" : "secondary"}>
                  {plugin.enabled ? "Enabled" : "Disabled"}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground">v{plugin.version} by {plugin.author}</p>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <p className="text-sm text-muted-foreground">{plugin.description}</p>
                
                <div className="text-xs text-muted-foreground">
                  <p>Created: {new Date(plugin.created_at).toLocaleString()}</p>
                  <p>Updated: {new Date(plugin.updated_at).toLocaleString()}</p>
                </div>

                {Object.keys(plugin.config).length > 0 && (
                  <div>
                    <p className="text-sm font-medium mb-1">Configuration:</p>
                    <div className="bg-gray-100 p-2 rounded text-xs max-h-20 overflow-auto">
                      <pre>{JSON.stringify(plugin.config, null, 2)}</pre>
                    </div>
                  </div>
                )}

                <div className="flex gap-2 pt-2">
                  <Button
                    size="sm"
                    variant={plugin.enabled ? "outline" : "default"}
                    onClick={() => handleTogglePlugin(plugin.id, !plugin.enabled)}
                    disabled={actionLoading !== null}
                  >
                    {actionLoading === `toggle-${plugin.id}` ? 'Processing...' : 
                     plugin.enabled ? (
                      <>
                        <PowerOff className="size-3 mr-1" />
                        Disable
                      </>
                    ) : (
                      <>
                        <Power className="size-3 mr-1" />
                        Enable
                      </>
                    )}
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleUninstallPlugin(plugin.id)}
                    disabled={actionLoading !== null}
                  >
                    {actionLoading === `uninstall-${plugin.id}` ? 'Uninstalling...' : (
                      <>
                        <Trash2 className="size-3 mr-1" />
                        Uninstall
                      </>
                    )}
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
        
        {filteredPlugins.length === 0 && (
          <div className="col-span-full text-center py-8">
            <p className="text-muted-foreground">No plugins found</p>
          </div>
        )}
      </div>
    </div>
  );
}
