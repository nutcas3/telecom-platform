import logging
from typing import Optional

from .auth import AuthProvider
from .api import (
    APIClient,
    SubscriberAPI,
    UsageAPI,
    PaymentAPI,
    RatingPlanAPI,
    SystemAPI,
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


class TelecomSDK:
    """Main SDK client for the Telecom Platform."""
    
    def __init__(
        self,
        api_url: str = "http://localhost:8000",
        api_key: Optional[str] = None,
        jwt_secret: Optional[str] = None,
        timeout: int = 30,
        max_retries: int = 3,
        retry_delay: float = 1.0,
        enable_logging: bool = False,
    ):
        """
        Initialize the Telecom SDK client.
        
        Args:
            api_url: Base URL for the API server
            api_key: API key for authentication
            jwt_secret: Secret key for JWT token generation/validation
            timeout: Request timeout in seconds
            max_retries: Maximum number of retry attempts
            retry_delay: Delay between retries in seconds
            enable_logging: Enable debug logging
        """
        self.api_url = api_url.rstrip('/')
        self.enable_logging = enable_logging
        
        if enable_logging:
            logging.basicConfig(level=logging.DEBUG)
            self.logger = logging.getLogger(__name__)
        else:
            self.logger = logging.getLogger(__name__)
            self.logger.disabled = True
        
        # Initialize authentication provider
        self.auth_provider = AuthProvider(
            api_key=api_key,
            jwt_secret=jwt_secret,
        )
        
        # Initialize API client
        self.api_client = APIClient(
            base_url=api_url,
            auth_provider=self.auth_provider,
            timeout=timeout,
            max_retries=max_retries,
            retry_delay=retry_delay,
            enable_logging=enable_logging,
        )
        
        # Initialize WebSocket client
        self.websocket_client = WebSocketClient(
            api_url=api_url,
            auth_provider=self.auth_provider,
            enable_logging=enable_logging,
        )
        
        # Initialize API modules
        self.subscribers = SubscriberAPI(self.api_client)
        self.usage = UsageAPI(self.api_client)
        self.payments = PaymentAPI(self.api_client)
        self.rating_plans = RatingPlanAPI(self.api_client)
        self.system = SystemAPI(self.api_client)
        self.analytics = AnalyticsAPI(self.api_client)
        self.security = SecurityAPI(self.api_client)
        self.currency = CurrencyAPI(self.api_client)
    
    async def __aenter__(self):
        """Async context manager entry."""
        await self.api_client._ensure_session()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.close()
    
    # Authentication methods
    
    def generate_jwt_token(
        self,
        user_id: str,
        expiry_hours: int = 24,
        **claims
    ) -> str:
        """Generate a JWT token for authentication."""
        return self.auth_provider.generate_jwt_token(user_id, expiry_hours, **claims)
    
    def validate_jwt_token(self, token: str) -> dict:
        """Validate a JWT token."""
        return self.auth_provider.validate_jwt_token(token)
    
    # WebSocket methods
    
    async def connect_websocket(self, message_handler) -> None:
        """Connect to WebSocket for real-time updates."""
        await self.websocket_client.connect(message_handler)
    
    async def disconnect_websocket(self) -> None:
        """Disconnect from WebSocket."""
        await self.websocket_client.disconnect()
    
    # Utility methods
    
    async def close(self) -> None:
        """Close the SDK client and cleanup resources."""
        await self.websocket_client.disconnect()
        await self.api_client.close()
        self.logger.info("Telecom SDK client closed")
