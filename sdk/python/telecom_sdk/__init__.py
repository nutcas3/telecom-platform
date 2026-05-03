from .client import TelecomSDK
from .auth import AuthProvider
from .api import (
    APIClient,
    SubscriberAPI,
    UsageAPI,
    PaymentAPI,
    RatingPlanAPI,
    SystemAPI,
    GraphQLAPI,
)
from .analytics import AnalyticsAPI
from .security import SecurityAPI
from .currency import CurrencyAPI
from .websocket import WebSocketClient
from .types import (
    Subscriber,
    UsageStats,
    PaymentTransaction,
    SystemStats,
    HealthStatus,
    WebSocketMessage,
    UsageEvent,
    RatingPlan,
    RealTimeUsage,
    PaginatedResponse,
    ChurnRiskLevel,
    ChurnPrediction,
    ChurnMetrics,
    FraudType,
    FraudSeverity,
    FraudAlert,
    FraudMetrics,
    FraudAlertFilter,
    MarketMetrics,
    PredictiveMaintenanceMetrics,
    PricingOptimizationResult,
    PricingMetrics,
)
from .exceptions import (
    TelecomError,
    AuthenticationError,
    APIError,
    NetworkError,
    ValidationError,
    RateLimitError,
    ServerError,
)

__version__ = "2.0.0"
__all__ = [
    # Main SDK client
    "TelecomSDK",
    # Authentication
    "AuthProvider",
    # API modules
    "APIClient",
    "SubscriberAPI",
    "UsageAPI",
    "PaymentAPI",
    "RatingPlanAPI",
    "SystemAPI",
    "GraphQLAPI",
    "AnalyticsAPI",
    "SecurityAPI",
    "CurrencyAPI",
    # WebSocket
    "WebSocketClient",
    # Types
    "Subscriber",
    "UsageStats",
    "PaymentTransaction",
    "SystemStats",
    "HealthStatus",
    "WebSocketMessage",
    "UsageEvent",
    "RatingPlan",
    "RealTimeUsage",
    "PaginatedResponse",
    "ChurnRiskLevel",
    "ChurnPrediction",
    "ChurnMetrics",
    "FraudType",
    "FraudSeverity",
    "FraudAlert",
    "FraudMetrics",
    "FraudAlertFilter",
    "MarketMetrics",
    "PredictiveMaintenanceMetrics",
    "PricingOptimizationResult",
    "PricingMetrics",
    # Exceptions
    "TelecomError",
    "AuthenticationError",
    "APIError",
    "NetworkError",
    "ValidationError",
    "RateLimitError",
    "ServerError",
]
