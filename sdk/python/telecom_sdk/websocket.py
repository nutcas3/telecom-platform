import json
import logging
from typing import Optional, Callable
import websockets
from websockets.exceptions import ConnectionClosed

from .types import WebSocketMessage
from .exceptions import NetworkError


class WebSocketClient:
    """WebSocket client for real-time updates from the Telecom Platform."""
    
    def __init__(
        self,
        api_url: str,
        auth_provider,
        enable_logging: bool = False,
    ):
        """
        Initialize the WebSocket client.
        
        Args:
            api_url: Base URL for the API server
            auth_provider: Authentication provider instance
            enable_logging: Enable debug logging
        """
        self.api_url = api_url.rstrip('/')
        self.auth_provider = auth_provider
        self.enable_logging = enable_logging
        
        if enable_logging:
            self.logger = logging.getLogger(__name__)
        else:
            self.logger = logging.getLogger(__name__)
            self.logger.disabled = True
        
        self._websocket: Optional[websockets.WebSocketServerProtocol] = None
        self._connected = False
    
    async def connect(self, message_handler: Callable[[WebSocketMessage], None]) -> None:
        """
        Connect to WebSocket for real-time updates.
        
        Args:
            message_handler: Async function to handle WebSocket messages
        """
        ws_url = self.api_url.replace("http://", "ws://") + "/ws"
        headers = self.auth_provider.get_headers()
        
        try:
            self._websocket = await websockets.connect(ws_url, extra_headers=headers)
            self._connected = True
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
            self._connected = False
        except Exception as e:
            self.logger.error(f"WebSocket error: {e}")
            self._connected = False
            raise NetworkError(f"WebSocket connection failed: {e}")
    
    async def disconnect(self) -> None:
        """Disconnect from WebSocket."""
        if self._websocket:
            await self._websocket.close()
            self._websocket = None
            self._connected = False
            self.logger.info("WebSocket disconnected")
    
    def is_connected(self) -> bool:
        """Check if WebSocket is connected."""
        return self._connected
    
    async def send(self, data: dict) -> None:
        """
        Send data through the WebSocket connection.
        
        Args:
            data: Dictionary data to send
        """
        if not self._connected or not self._websocket:
            raise NetworkError("WebSocket is not connected")
        
        try:
            await self._websocket.send(json.dumps(data))
        except Exception as e:
            self.logger.error(f"Failed to send WebSocket message: {e}")
            raise NetworkError(f"Failed to send message: {e}")
