"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Rocket, RotateCcw, Play, History, Clock, CheckCircle, XCircle, AlertCircle } from "lucide-react";
import { api, Deployment } from "@/lib/api";

export default function DeployPage() {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showDeployForm, setShowDeployForm] = useState(false);
  const [deployForm, setDeployForm] = useState({
    service: "",
    version: "",
    metadata: "{}"
  });
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    loadDeployments();
  }, []);

  const loadDeployments = async () => {
    try {
      setLoading(true);
      const deploymentsResponse = await api.deployments.status();
      setDeployments(deploymentsResponse.deployments);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load deployments");
    } finally {
      setLoading(false);
    }
  };

  const handleStartDeployment = async () => {
    try {
      setActionLoading('deploy');
      const metadata = deployForm.metadata ? JSON.parse(deployForm.metadata) : {};
      await api.deployments.start({
        service: deployForm.service,
        version: deployForm.version,
        metadata
      });
      
      // Reset form
      setDeployForm({ service: "", version: "", metadata: "{}" });
      setShowDeployForm(false);
      await loadDeployments();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start deployment");
    } finally {
      setActionLoading(null);
    }
  };

  const handleRollback = async (deploymentId: number) => {
    try {
      setActionLoading(`rollback-${deploymentId}`);
      await api.deployments.rollback(deploymentId);
      await loadDeployments();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to rollback deployment");
    } finally {
      setActionLoading(null);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircle className="size-4 text-green-500" />;
      case 'failed': return <XCircle className="size-4 text-red-500" />;
      case 'running': return <Play className="size-4 text-blue-500" />;
      case 'pending': return <Clock className="size-4 text-yellow-500" />;
      case 'rolling_back': return <RotateCcw className="size-4 text-orange-500" />;
      default: return <AlertCircle className="size-4 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'success';
      case 'failed': return 'destructive';
      case 'running': return 'default';
      case 'pending': return 'secondary';
      case 'rolling_back': return 'warning';
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
          <h3 className="text-red-800 font-medium">Error loading deployments</h3>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Rocket className="size-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Deployments</h1>
            <p className="text-muted-foreground mt-1">Deployment pipeline management and history.</p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button onClick={loadDeployments} disabled={loading}>
            <RotateCcw className={`size-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button onClick={() => setShowDeployForm(true)}>
            <Play className="size-4 mr-2" />
            New Deployment
          </Button>
        </div>
      </div>

      {/* Deployment Form */}
      {showDeployForm && (
        <Card>
          <CardHeader>
            <CardTitle>Start New Deployment</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium">Service Name</label>
                <input
                  type="text"
                  value={deployForm.service}
                  onChange={(e) => setDeployForm({ ...deployForm, service: e.target.value })}
                  placeholder="e.g., api-server"
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Version</label>
                <input
                  type="text"
                  value={deployForm.version}
                  onChange={(e) => setDeployForm({ ...deployForm, version: e.target.value })}
                  placeholder="e.g., v1.2.3"
                  className="mt-1 block w-full h-9 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Metadata (JSON)</label>
                <textarea
                  value={deployForm.metadata}
                  onChange={(e) => setDeployForm({ ...deployForm, metadata: e.target.value })}
                  placeholder='{"commit": "abc123", "author": "John Doe"}'
                  className="mt-1 block w-full h-20 rounded-lg border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring/50"
                />
              </div>
              <div className="flex gap-2">
                <Button onClick={handleStartDeployment} disabled={actionLoading !== null || !deployForm.service || !deployForm.version}>
                  {actionLoading === 'deploy' ? 'Deploying...' : 'Start Deployment'}
                </Button>
                <Button variant="outline" onClick={() => setShowDeployForm(false)}>
                  Cancel
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Deployment Statistics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <CheckCircle className="size-5 text-green-500" />
              <div>
                <p className="text-2xl font-bold">
                  {deployments.filter(d => d.status === 'completed').length}
                </p>
                <p className="text-sm text-muted-foreground">Completed</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <Play className="size-5 text-blue-500" />
              <div>
                <p className="text-2xl font-bold">
                  {deployments.filter(d => d.status === 'running').length}
                </p>
                <p className="text-sm text-muted-foreground">Running</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <XCircle className="size-5 text-red-500" />
              <div>
                <p className="text-2xl font-bold">
                  {deployments.filter(d => d.status === 'failed').length}
                </p>
                <p className="text-sm text-muted-foreground">Failed</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <History className="size-5 text-gray-500" />
              <div>
                <p className="text-2xl font-bold">{deployments.length}</p>
                <p className="text-sm text-muted-foreground">Total</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Deployments List */}
      <Card>
        <CardHeader>
          <CardTitle>Deployment History</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {deployments.map((deployment) => (
              <div key={deployment.id} className="border rounded-lg p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    {getStatusIcon(deployment.status)}
                    <div>
                      <h3 className="font-medium">{deployment.service}</h3>
                      <p className="text-sm text-muted-foreground">Version {deployment.version}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={getStatusColor(deployment.status)}>
                      {deployment.status}
                    </Badge>
                    <span className="text-sm text-muted-foreground">
                      ID: {deployment.id}
                    </span>
                  </div>
                </div>
                
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div>
                    <p className="text-muted-foreground">Started</p>
                    <p className="font-medium">{new Date(deployment.started_at).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Duration</p>
                    <p className="font-medium">{formatDuration(deployment.started_at, deployment.completed_at)}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Completed</p>
                    <p className="font-medium">
                      {deployment.completed_at ? new Date(deployment.completed_at).toLocaleString() : 'In progress'}
                    </p>
                  </div>
                  <div className="flex items-end">
                    {deployment.status === 'completed' && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleRollback(deployment.id)}
                        disabled={actionLoading !== null}
                      >
                        {actionLoading === `rollback-${deployment.id}` ? 'Rolling back...' : 'Rollback'}
                      </Button>
                    )}
                  </div>
                </div>

                {deployment.metadata && Object.keys(deployment.metadata).length > 0 && (
                  <div className="mt-3 pt-3 border-t">
                    <p className="text-sm text-muted-foreground mb-2">Metadata:</p>
                    <div className="bg-gray-100 p-2 rounded text-xs">
                      <pre>{JSON.stringify(deployment.metadata, null, 2)}</pre>
                    </div>
                  </div>
                )}
              </div>
            ))}
            
            {deployments.length === 0 && (
              <p className="text-muted-foreground text-center py-8">No deployments found</p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
