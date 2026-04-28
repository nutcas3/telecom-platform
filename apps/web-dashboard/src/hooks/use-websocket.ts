import { useEffect, useState, useRef, useCallback } from "react";
import { getWebSocketService, WebSocketMessage, ServiceUpdate, SystemMetrics, AlertUpdate } from "@/lib/websocket";

export function useWebSocket<T = any>(messageType: string) {
  const [data, setData] = useState<T | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const unsubscribeRef = useRef<(() => void) | null>(null);

  useEffect(() => {
    const wsService = getWebSocketService();

    // Subscribe to the specific message type
    const unsubscribe = wsService.subscribe(messageType, (messageData: T) => {
      setData(messageData);
      setError(null);
    });

    unsubscribeRef.current = unsubscribe;

    // Set up connection status monitoring
    const checkConnection = () => {
      setIsConnected(wsService.isConnected());
    };

    const interval = setInterval(checkConnection, 1000);

    return () => {
      if (unsubscribeRef.current) {
        unsubscribeRef.current();
      }
      clearInterval(interval);
    };
  }, [messageType]);

  return { data, isConnected, error };
}

export function useServiceUpdates() {
  return useWebSocket<ServiceUpdate>('service_update');
}

export function useSystemMetrics() {
  return useWebSocket<SystemMetrics>('system_metrics');
}

export function useAlertUpdates() {
  return useWebSocket<AlertUpdate>('alert_update');
}

export function useWebSocketConnection() {
  const [isConnected, setIsConnected] = useState(false);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);

  useEffect(() => {
    const wsService = getWebSocketService();

    const checkConnection = () => {
      setIsConnected(wsService.isConnected());
    };

    const interval = setInterval(checkConnection, 1000);

    return () => {
      clearInterval(interval);
    };
  }, []);

  const reconnect = useCallback(() => {
    // Force reconnection by disconnecting and reconnecting
    const wsService = getWebSocketService();
    wsService.disconnect();
    setTimeout(() => {
      const newService = getWebSocketService();
      setReconnectAttempts(prev => prev + 1);
    }, 1000);
  }, []);

  return { isConnected, reconnect, reconnectAttempts };
}
