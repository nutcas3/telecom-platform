import asyncio
import logging
from typing import Optional, Dict, Any, List
from datetime import datetime
import aiohttp

from .auth import AuthProvider
from .exceptions import (
    TelecomError,
    AuthenticationError,
    APIError,
    NetworkError,
    RateLimitError,
    ServerError,
)
from .types import (
    Subscriber,
    UsageStats,
    PaymentTransaction,
    SystemStats,
    HealthStatus,
    RealTimeUsage,
    PaginatedResponse,
    RatingPlan,
)


class APIClient:
    """HTTP client for making API requests to the Telecom Platform."""
    
    def __init__(
        self,
        base_url: str,
        auth_provider: AuthProvider,
        timeout: int = 30,
        max_retries: int = 3,
        retry_delay: float = 1.0,
        enable_logging: bool = False,
    ):
        """
        Initialize the API client.
        
        Args:
            base_url: Base URL for the API server
            auth_provider: Authentication provider instance
            timeout: Request timeout in seconds
            max_retries: Maximum number of retry attempts
            retry_delay: Delay between retries in seconds
            enable_logging: Enable debug logging
        """
        self.base_url = base_url.rstrip('/')
        self.auth_provider = auth_provider
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
            headers = self.auth_provider.get_headers()
            self._session = aiohttp.ClientSession(
                headers=headers,
                timeout=self.timeout,
            )
    
    async def make_request(
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
        url = f"{self.base_url}{endpoint}"
        
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
    
    async def close(self) -> None:
        """Close the HTTP session."""
        if self._session and not self._session.closed:
            await self._session.close()
        self.logger.info("API client closed")


class SubscriberAPI:
    """API endpoints for subscriber management."""
    
    def __init__(self, client: APIClient):
        self.client = client
    
    async def get(self, subscriber_id: int) -> Subscriber:
        """Get subscriber by ID."""
        response = await self.client.make_request("GET", f"/v1/subscribers/{subscriber_id}")
        return Subscriber(**response)
    
    async def list(
        self,
        page: int = 1,
        page_size: int = 50,
        status: Optional[str] = None,
    ) -> PaginatedResponse:
        """List subscribers with pagination."""
        params = {"page": page, "page_size": page_size}
        if status:
            params["status"] = status
        
        response = await self.client.make_request("GET", "/v1/subscribers", params=params)
        return PaginatedResponse(**response)
    
    async def create(
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
        
        response = await self.client.make_request("POST", "/v1/subscribers", data=data)
        return Subscriber(**response)
    
    async def update(self, subscriber_id: int, **kwargs) -> Subscriber:
        """Update subscriber details."""
        response = await self.client.make_request(
            "PUT",
            f"/v1/subscribers/{subscriber_id}",
            data=kwargs,
        )
        return Subscriber(**response)
    
    async def suspend(self, subscriber_id: int) -> Subscriber:
        """Suspend a subscriber."""
        response = await self.client.make_request("POST", f"/v1/subscribers/{subscriber_id}/suspend")
        return Subscriber(**response)
    
    async def activate(self, subscriber_id: int) -> Subscriber:
        """Activate a suspended subscriber."""
        response = await self.client.make_request("POST", f"/v1/subscribers/{subscriber_id}/activate")
        return Subscriber(**response)
    
    async def terminate(self, subscriber_id: int) -> bool:
        """Terminate a subscriber."""
        response = await self.client.make_request("DELETE", f"/v1/subscribers/{subscriber_id}")
        return response.get("success", False)


class UsageAPI:
    """API endpoints for usage management."""
    
    def __init__(self, client: APIClient):
        self.client = client
    
    async def get_stats(
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
        response = await self.client.make_request(
            "GET",
            f"/v1/subscribers/{subscriber_id}/usage",
            params=params,
        )
        return UsageStats(**response)
    
    async def list_events(
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
        
        response = await self.client.make_request("GET", "/v1/usage/events", params=params)
        return PaginatedResponse(**response)
    
    async def get_realtime(self, subscriber_id: int) -> RealTimeUsage:
        """Get real-time usage for a subscriber."""
        response = await self.client.make_request("GET", f"/v1/subscribers/{subscriber_id}/realtime")
        return RealTimeUsage(**response)


class PaymentAPI:
    """API endpoints for payment management."""
    
    def __init__(self, client: APIClient):
        self.client = client
    
    async def create_transaction(
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
        
        response = await self.client.make_request("POST", "/v1/payments/transactions", data=data)
        return PaymentTransaction(**response)
    
    async def get_transaction(self, transaction_id: str) -> PaymentTransaction:
        """Get payment transaction by ID."""
        response = await self.client.make_request("GET", f"/v1/payments/transactions/{transaction_id}")
        return PaymentTransaction(**response)
    
    async def list_transactions(
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
        
        response = await self.client.make_request("GET", "/v1/payments/transactions", params=params)
        return PaginatedResponse(**response)


class RatingPlanAPI:
    """API endpoints for rating plan management."""
    
    def __init__(self, client: APIClient):
        self.client = client
    
    async def list(self) -> List[RatingPlan]:
        """List all available rating plans."""
        response = await self.client.make_request("GET", "/v1/rating-plans")
        return [RatingPlan(**plan) for plan in response]
    
    async def get(self, plan_id: str) -> RatingPlan:
        """Get rating plan by ID."""
        response = await self.client.make_request("GET", f"/v1/rating-plans/{plan_id}")
        return RatingPlan(**response)


class SystemAPI:
    """API endpoints for system management."""
    
    def __init__(self, client: APIClient):
        self.client = client
    
    async def get_stats(self) -> SystemStats:
        """Get system statistics."""
        response = await self.client.make_request("GET", "/v1/system/stats")
        return SystemStats(**response)
    
    async def get_health(self) -> HealthStatus:
        """Get system health status."""
        response = await self.client.make_request("GET", "/v1/health")
        return HealthStatus(**response)


class GraphQLAPI:
    """API endpoints for GraphQL queries."""
    
    def __init__(self, client: APIClient):
        self.client = client
    
    async def execute(
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
        
        response = await self.client.make_request("POST", "/graphql", data=data)
        return response
