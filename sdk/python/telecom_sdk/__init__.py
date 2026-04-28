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

__version__ = "1.0.0"
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
    # Exceptions
    "TelecomError",
    "AuthenticationError",
    "APIError",
    "NetworkError",
    "ValidationError",
    "RateLimitError",
    "ServerError",
]
