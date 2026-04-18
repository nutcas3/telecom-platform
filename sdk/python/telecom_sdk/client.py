import asyncio
import json
import logging
from typing import Optional, Dict, Any, List, Union
from datetime import datetime
import aiohttp
import websockets
from websockets.exceptions import ConnectionClosed

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
    
    def __init__(
        self,
        api_url: str = "http://localhost:8000",
        api_key: Optional[str] = None,
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
            timeout: Request timeout in seconds
            max_retries: Maximum number of retry attempts
            retry_delay: Delay between retries in seconds
            enable_logging: Enable debug logging
        """
        self.api_url = api_url.rstrip('/')
        self.api_key = api_key
        self.timeout = aiohttp.ClientTimeout(total=timeout)
        self.max_retries = max_retries
        self.retry_delay = retry_delay
        self.enable_logging = enable_logging
        
        if enable_logging:
            logging.basicConfig(level=logging.DEBUG)
            self.logger = logging.getLogger(__name__)
        else:
            self.logger = logging.getLogger(__name__)
            self.logger.disabled = True
        
        self._session: Optional[aiohttp.ClientSession] = None
        self._websocket: Optional[websockets.WebSocketServerProtocol] = None
    
    async def __aenter__(self):
        """Async context manager entry."""
        await self._ensure_session()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.close()
    
    async def _ensure_session(self):
        """Ensure aiohttp session is created."""
        if self._session is None or self._session.closed:
            headers = self._get_headers()
            self._session = aiohttp.ClientSession(
                headers=headers,
                timeout=self.timeout,
            )
    
    def _get_headers(self) -> Dict[str, str]:
        """Get default headers for API requests."""
        headers = {
            "Content-Type": "application/json",
            "User-Agent": f"Telecom-Python-SDK/1.0.0",
        }
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        return headers
    
    async def _make_request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, str]] = None,
    ) -> Dict[str, Any]:
        """
        Make an HTTP request with retry logic.
        
        Args:
            method: HTTP method
            endpoint: API endpoint
            data: Request body data
            params: Query parameters
            
        Returns:
            Response data as dictionary
            
        Raises:
            TelecomError: On API errors
        """
        await self._ensure_session()
        url = f"{self.api_url}{endpoint}"
        
        for attempt in range(self.max_retries + 1):
            try:
                self.logger.debug(f"Making {method} request to {url} (attempt {attempt + 1})")
                
                async with self._session.request(
                    method=method,
                    url=url,
                    json=data,
                    params=params,
                ) as response:
                    await self._handle_response_errors(response)
                    return await response.json()
                    
            except aiohttp.ClientError as e:
                if attempt == self.max_retries:
                    raise NetworkError(f"Network error after {self.max_retries} retries: {e}")
                
                self.logger.warning(f"Request failed (attempt {attempt + 1}), retrying in {self.retry_delay}s: {e}")
                await asyncio.sleep(self.retry_delay * (2 ** attempt))
    
    async def _handle_response_errors(self, response: aiohttp.ClientResponse):
        """Handle HTTP response errors."""
        if response.status == 401:
            raise AuthenticationError("Authentication failed")
        elif response.status == 429:
            raise RateLimitError("Rate limit exceeded")
        elif 400 <= response.status < 500:
            error_data = await response.json().get("error", "Bad request")
            raise APIError(f"API error: {error_data}", response.status)
        elif response.status >= 500:
            raise ServerError(f"Server error: {response.status}", response.status)
    
    # Subscriber Management
    
    async def get_subscriber(self, subscriber_id: int) -> Subscriber:
        """Get subscriber by ID."""
        response = await self._make_request("GET", f"/v1/subscribers/{subscriber_id}")
        return Subscriber(**response)
    
    async def list_subscribers(
        self,
        page: int = 1,
        page_size: int = 50,
        status: Optional[str] = None,
    ) -> PaginatedResponse:
        """List subscribers with pagination."""
        params = {"page": page, "page_size": page_size}
        if status:
            params["status"] = status
        
        response = await self._make_request("GET", "/v1/subscribers", params=params)
        return PaginatedResponse(**response)
    
    async def create_subscriber(
        self,
        imsi: str,
        msisdn: str,
        first_name: str,
        last_name: str,
        email: str,
        plan_id: int,
        organization_id: Optional[str] = None,
    ) -> Subscriber:
        """Create a new subscriber."""
        data = {
            "imsi": imsi,
            "msisdn": msisdn,
            "first_name": first_name,
            "last_name": last_name,
            "email": email,
            "plan_id": plan_id,
        }
        if organization_id:
            data["organization_id"] = organization_id
        
        response = await self._make_request("POST", "/v1/subscribers", data=data)
        return Subscriber(**response)
    
    async def update_subscriber(
        self,
        subscriber_id: int,
        **kwargs,
    ) -> Subscriber:
        """Update subscriber details."""
        response = await self._make_request(
            "PUT",
            f"/v1/subscribers/{subscriber_id}",
            data=kwargs,
        )
        return Subscriber(**response)
    
    async def suspend_subscriber(self, subscriber_id: int) -> Subscriber:
        """Suspend a subscriber."""
        response = await self._make_request("POST", f"/v1/subscribers/{subscriber_id}/suspend")
        return Subscriber(**response)
    
    async def activate_subscriber(self, subscriber_id: int) -> Subscriber:
        """Activate a suspended subscriber."""
        response = await self._make_request("POST", f"/v1/subscribers/{subscriber_id}/activate")
        return Subscriber(**response)
    
    async def terminate_subscriber(self, subscriber_id: int) -> bool:
        """Terminate a subscriber."""
        response = await self._make_request("DELETE", f"/v1/subscribers/{subscriber_id}")
        return response.get("success", False)
    
    # Usage Management
    
    async def get_usage_stats(
        self,
        subscriber_id: int,
        start_date: datetime,
        end_date: datetime,
    ) -> UsageStats:
        """Get usage statistics for a subscriber."""
        params = {
            "start_date": start_date.isoformat(),
            "end_date": end_date.isoformat(),
        }
        response = await self._make_request(
            "GET",
            f"/v1/subscribers/{subscriber_id}/usage",
            params=params,
        )
        return UsageStats(**response)
    
    async def list_usage_events(
        self,
        subscriber_id: Optional[int] = None,
        usage_type: Optional[str] = None,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        page: int = 1,
        page_size: int = 50,
    ) -> PaginatedResponse:
        """List usage events with filtering."""
        params = {"page": page, "page_size": page_size}
        
        if subscriber_id:
            params["subscriber_id"] = subscriber_id
        if usage_type:
            params["usage_type"] = usage_type
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()
        
        response = await self._make_request("GET", "/v1/usage/events", params=params)
        return PaginatedResponse(**response)
    
    async def get_real_time_usage(self, subscriber_id: int) -> RealTimeUsage:
        """Get real-time usage for a subscriber."""
        response = await self._make_request("GET", f"/v1/subscribers/{subscriber_id}/realtime")
        return RealTimeUsage(**response)
    
    # Payment Management
    
    async def create_payment_transaction(
        self,
        subscriber_id: int,
        amount: float,
        currency: str = "USD",
        gateway: str = "stripe",
        metadata: Optional[Dict[str, Any]] = None,
    ) -> PaymentTransaction:
        """Create a payment transaction."""
        data = {
            "subscriber_id": subscriber_id,
            "amount": amount,
            "currency": currency,
            "gateway": gateway,
        }
        if metadata:
            data["metadata"] = metadata
        
        response = await self._make_request("POST", "/v1/payments/transactions", data=data)
        return PaymentTransaction(**response)
    
    async def get_payment_transaction(self, transaction_id: str) -> PaymentTransaction:
        """Get payment transaction by ID."""
        response = await self._make_request("GET", f"/v1/payments/transactions/{transaction_id}")
        return PaymentTransaction(**response)
    
    async def list_payment_transactions(
        self,
        subscriber_id: Optional[int] = None,
        status: Optional[str] = None,
        page: int = 1,
        page_size: int = 50,
    ) -> PaginatedResponse:
        """List payment transactions with filtering."""
        params = {"page": page, "page_size": page_size}
        
        if subscriber_id:
            params["subscriber_id"] = subscriber_id
        if status:
            params["status"] = status
        
        response = await self._make_request("GET", "/v1/payments/transactions", params=params)
        return PaginatedResponse(**response)
    
    # Rating Plans
    
    async def list_rating_plans(self) -> List[RatingPlan]:
        """List all available rating plans."""
        response = await self._make_request("GET", "/v1/rating-plans")
        return [RatingPlan(**plan) for plan in response]
    
    async def get_rating_plan(self, plan_id: str) -> RatingPlan:
        """Get rating plan by ID."""
        response = await self._make_request("GET", f"/v1/rating-plans/{plan_id}")
        return RatingPlan(**response)
    
    # System Management
    
    async def get_system_stats(self) -> SystemStats:
        """Get system statistics."""
        response = await self._make_request("GET", "/v1/system/stats")
        return SystemStats(**response)
    
    async def get_health_status(self) -> HealthStatus:
        """Get system health status."""
        response = await self._make_request("GET", "/v1/health")
        return HealthStatus(**response)
    
    # WebSocket Support
    
    async def connect_websocket(self, message_handler: callable) -> None:
        """
        Connect to WebSocket for real-time updates.
        
        Args:
            message_handler: Async function to handle WebSocket messages
        """
        ws_url = self.api_url.replace("http://", "ws://") + "/ws"
        
        try:
            self._websocket = await websockets.connect(ws_url)
            self.logger.info("WebSocket connected")
            
            async for message in self._websocket:
                try:
                    data = json.loads(message)
                    ws_message = WebSocketMessage(**data)
                    await message_handler(ws_message)
                except json.JSONDecodeError as e:
                    self.logger.error(f"Failed to decode WebSocket message: {e}")
                except Exception as e:
                    self.logger.error(f"Error handling WebSocket message: {e}")
                    
        except ConnectionClosed:
            self.logger.warning("WebSocket connection closed")
        except Exception as e:
            self.logger.error(f"WebSocket error: {e}")
            raise NetworkError(f"WebSocket connection failed: {e}")
    
    async def disconnect_websocket(self) -> None:
        """Disconnect from WebSocket."""
        if self._websocket:
            await self._websocket.close()
            self._websocket = None
            self.logger.info("WebSocket disconnected")
    
    # GraphQL Support
    
    async def execute_graphql_query(
        self,
        query: str,
        variables: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """
        Execute a GraphQL query.
        
        Args:
            query: GraphQL query string
            variables: Optional variables for the query
            
        Returns:
            GraphQL response data
        """
        data = {"query": query}
        if variables:
            data["variables"] = variables
        
        response = await self._make_request("POST", "/graphql", data=data)
        return response
    
    # Utility Methods
    
    async def close(self) -> None:
        """Close the SDK client and cleanup resources."""
        if self._websocket:
            await self.disconnect_websocket()
        
        if self._session and not self._session.closed:
            await self._session.close()
        
        self.logger.info("Telecom SDK client closed")
