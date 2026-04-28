"use client";

import { Loader2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

interface LoadingStateProps {
  loading: boolean;
  error?: string | null;
  onRetry?: () => void;
  retryText?: string;
  loadingText?: string;
  children: React.ReactNode;
  className?: string;
}

export function LoadingState({ 
  loading, 
  error, 
  onRetry, 
  retryText = "Retry",
  loadingText = "Loading...",
  children,
  className = ""
}: LoadingStateProps) {
  if (loading) {
    return (
      <Card className={className}>
        <CardContent className="flex flex-col items-center justify-center p-8">
          <Loader2 className="h-8 w-8 animate-spin text-blue-600 mb-4" />
          <p className="text-gray-600 text-center">{loadingText}</p>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className={className}>
        <CardContent className="flex flex-col items-center justify-center p-8">
          <div className="text-center">
            <p className="text-red-600 mb-4">{error}</p>
            {onRetry && (
              <Button
                variant="outline"
                onClick={onRetry}
                className="flex items-center gap-2"
              >
                <RefreshCw className="h-4 w-4" />
                {retryText}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    );
  }

  return <>{children}</>;
}

interface InlineLoadingProps {
  loading: boolean;
  text?: string;
  children: React.ReactNode;
}

export function InlineLoading({ loading, text = "Loading...", children }: InlineLoadingProps) {
  if (loading) {
    return (
      <div className="flex items-center gap-2 text-gray-600">
        <Loader2 className="h-4 w-4 animate-spin" />
        <span>{text}</span>
      </div>
    );
  }

  return <>{children}</>;
}

interface ActionLoadingProps {
  loading: boolean;
  action?: string;
  children: React.ReactNode;
}

export function ActionLoading({ loading, action = "processing", children }: ActionLoadingProps) {
  if (loading) {
    return (
      <Button disabled className="flex items-center gap-2">
        <Loader2 className="h-4 w-4 animate-spin" />
        {action}...
      </Button>
    );
  }

  return <>{children}</>;
}
