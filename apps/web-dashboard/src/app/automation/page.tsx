"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Bot, Play, Clock, CheckCircle, XCircle, AlertCircle, History, Plus, Calendar } from "lucide-react";
import { api, Automation, AutomationRun } from "@/lib/api";

export default function AutomationPage() {
  const [automations, setAutomations] = useState<Automation[]>([]);
  const [runs, setRuns] = useState<AutomationRun[]>([]);
  const [selectedAutomation, setSelectedAutomation] = useState<Automation | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [createForm, setCreateForm] = useState({
    name: "",
    description: "",
    trigger: "",
    schedule: "",
    config: "{}"
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [automationsResponse, runsResponse] = await Promise.all([
        api.automation.list(),
        api.automation.logs()
      ]);
      
      setAutomations(automationsResponse.automations);
      setRuns(runsResponse.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load automation data");
    } finally {
      setLoading(false);
    }
  };

  const handleCreateAutomation = async () => {
    try {
      setActionLoading('create');
      const config = createForm.config ? JSON.parse(createForm.config) : {};
      await api.automation.create({
        name: createForm.name,
        description: createForm.description,
        trigger: createForm.trigger,
        schedule: createForm.schedule,
        config,
        enabled: true
      });
      
      // Reset form
      setCreateForm({ name: "", description: "", trigger: "", schedule: "", config: "{}" });
      setShowCreateForm(false);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create automation");
    } finally {
      setActionLoading(null);
    }
  };

  const handleRunAutomation = async (automationId: number) => {
    try {
      setActionLoading(`run-${automationId}`);
      await api.automation.run(automationId);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to run automation");
    } finally {
      setActionLoading(null);
    }
  };

  const handleScheduleAutomation = async (automationId: number, schedule: string) => {
    try {
      setActionLoading(`schedule-${automationId}`);
      await api.automation.schedule(automationId, schedule);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to schedule automation");
    } finally {
      setActionLoading(null);
    }
  };

  const loadAutomationRuns = async (automationId: number) => {
    try {
      const runsResponse = await api.automation.logs(automationId);
      setRuns(runsResponse.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load automation runs");
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircle className="size-4 text-green-500" />;
      case 'failed': return <XCircle className="size-4 text-red-500" />;
      case 'running': return <Play className="size-4 text-blue-500" />;
      case 'pending': return <Clock className="size-4 text-yellow-500" />;
      default: return <AlertCircle className="size-4 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'success';
      case 'failed': return 'destructive';
      case 'running': return 'default';
      case 'pending': return 'secondary';
      default: return 'secondary';
    }
  };

  const formatDuration = (started: string, completed?: string) => {
    const start = new Date(started);
    const end = completed ? new Date(completed) : new Date();
    const duration = end.getTime() - start.getTime();
    const minutes = Math.floor(duration / 60000);
    const seconds = Math.floor((duration % 60000) / 1000);
    return `${minutes}m ${seconds}s`;
  };

  if (loading) {
    return (
      <div className="p-8 space-y-6">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-48 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-96"></div>
        </div>
        <div className="grid gap-6 lg:grid-cols-2">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="animate-pulse">
              <div className="h-64 bg-gray-200 rounded"></div>
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
          <h3 className="text-red-800 font-medium">Error loading automation data</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Bot className="size-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Automation</h1>
            <p className="text-muted-foreground mt-1">Workflow automation and scheduling.</p>
          </div>
        </div>
        <Button onClick={() => setShowCreateForm(true)}>
          <Plus className="size-4 mr-2" />
          Create Automation
        </Button>
      </div>

      {/* Create Automation Form */}
      {showCreateForm && (
        <Card>
          <CardHeader>
            <CardTitle>Create New Automation</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium">Name</label>
                <input
                  type="text"
                  value={createForm.name}
                  onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
                  placeholder="e.g., Daily Backup"
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Description</label>
                <input
                  type="text"
                  value={createForm.description}
                  onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
                  placeholder="e.g., Backup database daily at midnight"
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Trigger</label>
                <input
                  type="text"
                  value={createForm.trigger}
                  onChange={(e) => setCreateForm({ ...createForm, trigger: e.target.value })}
                  placeholder="e.g., manual, schedule, webhook"
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Schedule (cron)</label>
                <input
                  type="text"
                  value={createForm.schedule}
                  onChange={(e) => setCreateForm({ ...createForm, schedule: e.target.value })}
                  placeholder="e.g., 0 0 * * * (daily at midnight)"
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Configuration (JSON)</label>
                <textarea
                  value={createForm.config}
                  onChange={(e) => setCreateForm({ ...createForm, config: e.target.value })}
                  placeholder='{"backup_path": "/backups", "retention": "7d"}'
                  className="mt-1 block w-full h-20 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div className="flex gap-2">
                <Button onClick={handleCreateAutomation} disabled={actionLoading !== null || !createForm.name || !createForm.trigger}>
                  {actionLoading === 'create' ? 'Creating...' : 'Create Automation'}
                </Button>
                <Button variant="outline" onClick={() => setShowCreateForm(false)}>
                  Cancel
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Automation Statistics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <Bot className="size-5 text-blue-500" />
              <div>
                <p className="text-2xl font-bold">{automations.length}</p>
                <p className="text-sm text-muted-foreground">Total Automations</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <CheckCircle className="size-5 text-green-500" />
              <div>
                <p className="text-2xl font-bold">{automations.filter(a => a.enabled).length}</p>
                <p className="text-sm text-muted-foreground">Enabled</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <Play className="size-5 text-purple-500" />
              <div>
                <p className="text-2xl font-bold">{runs.filter(r => r.status === 'running').length}</p>
                <p className="text-sm text-muted-foreground">Running</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <History className="size-5 text-orange-500" />
              <div>
                <p className="text-2xl font-bold">{runs.length}</p>
                <p className="text-sm text-muted-foreground">Total Runs</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Automations List */}
        <Card>
          <CardHeader>
            <CardTitle>Automations</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {automations.map((automation) => (
                <div key={automation.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <div>
                      <h3 className="font-medium">{automation.name}</h3>
                      <p className="text-sm text-muted-foreground">{automation.description}</p>
                    </div>
                    <Badge variant={automation.enabled ? "success" : "secondary"}>
                      {automation.enabled ? "Enabled" : "Disabled"}
                    </Badge>
                  </div>
                  
                  <div className="text-sm text-muted-foreground mb-3">
                    <p>Trigger: {automation.trigger}</p>
                    {automation.schedule && <p>Schedule: {automation.schedule}</p>}
                  </div>

                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      onClick={() => handleRunAutomation(automation.id)}
                      disabled={actionLoading !== null}
                    >
                      {actionLoading === `run-${automation.id}` ? 'Running...' : (
                        <>
                          <Play className="size-3 mr-1" />
                          Run Now
                        </>
                      )}
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setSelectedAutomation(automation)}
                    >
                      <History className="size-3 mr-1" />
                      View Runs
                    </Button>
                    {automation.trigger === 'schedule' && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          const newSchedule = prompt("Enter new schedule (cron):", automation.schedule || "");
                          if (newSchedule) handleScheduleAutomation(automation.id, newSchedule);
                        }}
                        disabled={actionLoading !== null}
                      >
                        {actionLoading === `schedule-${automation.id}` ? 'Updating...' : (
                          <>
                            <Calendar className="size-3 mr-1" />
                            Schedule
                          </>
                        )}
                      </Button>
                    )}
                  </div>
                </div>
              ))}
              
              {automations.length === 0 && (
                <p className="text-muted-foreground text-center py-8">No automations found</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Runs History */}
        <Card>
          <CardHeader>
            <CardTitle>
              {selectedAutomation ? `${selectedAutomation.name} - Runs` : 'Recent Runs'}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {runs.map((run) => (
                <div key={run.id} className="border rounded p-3">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(run.status)}
                      <span className="font-medium text-sm">Run #{run.id}</span>
                    </div>
                    <Badge variant={getStatusColor(run.status)}>
                      {run.status}
                    </Badge>
                  </div>
                  
                  <div className="text-sm text-muted-foreground">
                    <p>Started: {new Date(run.started_at).toLocaleString()}</p>
                    <p>Duration: {formatDuration(run.started_at, run.completed_at)}</p>
                    {run.completed_at && (
                      <p>Completed: {new Date(run.completed_at).toLocaleString()}</p>
                    )}
                  </div>

                  {run.output && (
                    <div className="mt-2">
                      <p className="text-sm font-medium mb-1">Output:</p>
                      <div className="bg-gray-100 p-2 rounded text-xs max-h-20 overflow-auto">
                        <pre>{run.output}</pre>
                      </div>
                    </div>
                  )}

                  {run.error && (
                    <div className="mt-2">
                      <p className="text-sm font-medium mb-1 text-red-600">Error:</p>
                      <div className="bg-red-50 p-2 rounded text-xs text-red-700 max-h-20 overflow-auto">
                        <pre>{run.error}</pre>
                      </div>
                    </div>
                  )}
                </div>
              ))}
              
              {runs.length === 0 && (
                <p className="text-muted-foreground text-center py-8">No runs found</p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
