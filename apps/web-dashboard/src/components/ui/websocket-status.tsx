"use client";

import { Wifi, WifiOff, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useWebSocketConnection } from "@/hooks/use-websocket";

export function WebSocketStatus() {
  const { isConnected, reconnect, reconnectAttempts } = useWebSocketConnection();

  return (
    <div className="flex items-center gap-2">
      <Badge variant={isConnected ? "success" : "destructive"} className="flex items-center gap-1">
        {isConnected ? (
          <>
            <Wifi className="h-3 w-3" />
            Live
          </>
        ) : (
          <>
            <WifiOff className="h-3 w-3" />
            Offline
          </>
        )}
      </Badge>
      
      {!isConnected && (
        <Button
          variant="outline"
          size="sm"
          onClick={reconnect}
          className="flex items-center gap-1"
        >
          <RefreshCw className="h-3 w-3" />
          Reconnect
        </Button>
      )}
      
      {reconnectAttempts > 0 && (
        <span className="text-xs text-muted-foreground">
          Attempts: {reconnectAttempts}
        </span>
      )}
    </div>
  );
}
