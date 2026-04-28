"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Server, Play, Square, RefreshCw, Eye, Activity, AlertTriangle } from "lucide-react";
import { api, Service, ServiceHealth } from "@/lib/api";
import { ErrorAlert } from "@/components/ui/error-alert";
import { LoadingState, ActionLoading } from "@/components/ui/loading-state";
import { useServiceUpdates } from "@/hooks/use-websocket";
import { WebSocketStatus } from "@/components/ui/websocket-status";

export default function ServicesPage() {
  const [services, setServices] = useState<Service[]>([]);
  const [selectedService, setSelectedService] = useState<Service | null>(null);
  const [health, setHealth] = useState<ServiceHealth | null>(null);
  const [podStatus, setPodStatus] = useState<any>(null);
  const [logs, setLogs] = useState<string>("");
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  
  // WebSocket real-time updates
  const serviceUpdate = useServiceUpdates();

  useEffect(() => {
    loadServices();
  }, []);

  // Handle real-time service updates
  useEffect(() => {
    if (serviceUpdate.data) {
      setServices(prevServices => 
        prevServices.map(service => 
          service.name === serviceUpdate.data?.service 
            ? { ...service, ...serviceUpdate.data }
            : service
        )
      );
    }
  }, [serviceUpdate.data]);

  const loadServices = async () => {
    try {
      setLoading(true);
      setError(null);
      const servicesResponse = await api.services.list();
      setServices(servicesResponse.services);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to load services";
      setError(errorMessage);
      console.error("Failed to load services:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleServiceAction = async (serviceName: string, action: string) => {
    try {
      setActionLoading(action);
      setError(null);
      switch (action) {
        case 'restart':
          await api.services.restart(serviceName);
          break;
        case 'start':
          await api.services.start(serviceName);
          break;
        case 'stop':
          await api.services.stop(serviceName);
          break;
      }
      // Refresh services after action
      await loadServices();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : `Failed to ${action} service`;
      setError(errorMessage);
      console.error(`Failed to ${action} service ${serviceName}:`, err);
    } finally {
      setActionLoading(null);
    }
  };

  const handleScale = async (serviceName: string, replicas: number) => {
    try {
      setActionLoading('scale');
      setError(null);
      await api.services.scale(serviceName, replicas);
      await loadServices();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to scale service";
      setError(errorMessage);
      console.error(`Failed to scale service ${serviceName} to ${replicas} replicas:`, err);
    } finally {
      setActionLoading(null);
    }
  };

  const loadServiceDetails = async (serviceName: string) => {
    try {
      setError(null);
      const [serviceDetails, healthData, podData, logsData, eventsData] = await Promise.all([
        api.services.get(serviceName),
        api.services.health(serviceName),
        api.services.podStatus(serviceName),
        api.services.logs(serviceName),
        api.services.events(serviceName)
      ]);
      
      setSelectedService(serviceDetails);
      setHealth(healthData);
      setPodStatus(podData);
      setLogs(logsData);
      setEvents(eventsData.events);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load service details");
    }
  };

  if (loading && services.length === 0) {
    return (
      <div className="p-8 space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Server className="size-8 text-primary" />
            <div>
              <h1 className="text-3xl font-bold tracking-tight">Services</h1>
              <p className="text-muted-foreground mt-1">Kubernetes service management and monitoring.</p>
            </div>
          </div>
          <Button onClick={loadServices} disabled={loading}>
            <RefreshCw className={`size-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>

        <LoadingState loading={loading} onRetry={loadServices}>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="animate-pulse">
                <div className="h-48 bg-gray-200 rounded"></div>
              </div>
            ))}
          </div>
        </LoadingState>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Server className="size-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Services</h1>
            <p className="text-muted-foreground mt-1">Kubernetes service management and monitoring.</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <WebSocketStatus />
          <Button onClick={loadServices} disabled={loading}>
            <RefreshCw className={`size-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      <ErrorAlert 
        error={error} 
        onRetry={loadServices}
        onDismiss={() => setError(null)}
        retryText="Retry Loading"
      />

      <LoadingState loading={loading} error={error} onRetry={loadServices}>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {services.map((service) => (
            <Card key={service.name} className="cursor-pointer hover:shadow-md transition-shadow" 
                  onClick={() => loadServiceDetails(service.name)}>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg">{service.name}</CardTitle>
                  <Badge variant={
                    service.readyReplicas === service.replicas ? "success" :
                    service.readyReplicas > 0 ? "warning" : "destructive"
                  }>
                    {service.readyReplicas}/{service.replicas}
                  </Badge>
                </div>
                <p className="text-sm text-muted-foreground">{service.namespace}</p>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div>
                      <p className="text-muted-foreground">Ready</p>
                      <p className="font-medium">{service.readyReplicas}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">Available</p>
                      <p className="font-medium">{service.availableReplicas}</p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <ActionLoading loading={actionLoading === 'restart'} action="restarting">
                      <Button size="sm" variant="outline" 
                              onClick={(e) => { e.stopPropagation(); handleServiceAction(service.name, 'restart'); }}
                              disabled={actionLoading !== null}>
                        <RefreshCw className="size-4 mr-1" />
                        Restart
                      </Button>
                    </ActionLoading>
                    <ActionLoading loading={actionLoading === 'start'} action="starting">
                      <Button size="sm" variant="outline"
                              onClick={(e) => { e.stopPropagation(); handleServiceAction(service.name, 'start'); }}
                              disabled={actionLoading !== null}>
                        <Play className="size-4 mr-1" />
                        Start
                      </Button>
                    </ActionLoading>
                    <ActionLoading loading={actionLoading === 'stop'} action="stopping">
                      <Button size="sm" variant="outline"
                              onClick={(e) => { e.stopPropagation(); handleServiceAction(service.name, 'stop'); }}
                              disabled={actionLoading !== null}>
                        <Square className="size-4 mr-1" />
                        Stop
                      </Button>
                    </ActionLoading>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </LoadingState>

      {selectedService && (
        <div className="grid gap-6 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="size-5" />
                {selectedService.name} - Details
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Namespace</p>
                    <p className="font-medium">{selectedService.namespace}</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Created</p>
                    <p className="font-medium">{new Date(selectedService.creationTimestamp).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Replicas</p>
                    <p className="font-medium">{selectedService.replicas}</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Ready</p>
                    <p className="font-medium">{selectedService.readyReplicas}</p>
                  </div>
                </div>
                
                {health && (
                  <div>
                    <p className="text-sm text-muted-foreground mb-2">Health Status</p>
                    <Badge variant={
                      health.status === 'healthy' ? 'success' :
                      health.status === 'degraded' ? 'warning' : 'destructive'
                    }>
                      {health.status}
                    </Badge>
                  </div>
                )}

                <div>
                  <p className="text-sm text-muted-foreground mb-2">Scale Service</p>
                  <div className="flex gap-2">
                    <input 
                      type="number" 
                      min="0" 
                      defaultValue={selectedService.replicas}
                      className="w-20 px-2 py-1 border rounded"
                      id={`scale-${selectedService.name}`}
                    />
                    <ActionLoading loading={actionLoading === 'scale'} action="scaling">
                      <Button size="sm" onClick={() => {
                        const input = document.getElementById(`scale-${selectedService.name}`) as HTMLInputElement;
                        const replicas = parseInt(input.value);
                        handleScale(selectedService.name, replicas);
                      }}>
                        Scale
                      </Button>
                    </ActionLoading>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Eye className="size-5" />
                Logs & Events
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <p className="text-sm text-muted-foreground mb-2">Recent Logs</p>
                  <div className="bg-gray-900 text-green-400 p-3 rounded text-sm font-mono h-32 overflow-auto">
                    {logs || "No logs available"}
                  </div>
                </div>
                
                <div>
                  <p className="text-sm text-muted-foreground mb-2">Recent Events</p>
                  <div className="space-y-2 max-h-32 overflow-auto">
                    {events.length > 0 ? (
                      events.slice(0, 5).map((event, index) => (
                        <div key={index} className="text-sm border-l-2 border-blue-500 pl-2">
                          <span className="text-muted-foreground">{event.type}</span>
                          <span className="ml-2">{event.message}</span>
                        </div>
                      ))
                    ) : (
                      <p className="text-sm text-muted-foreground">No recent events</p>
                    )}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
