class TelecomError(Exception):
    """Base exception for all Telecom SDK errors."""
    
    def __init__(self, message: str, status_code: Optional[int] = None, response_data: Optional[dict] = None):
        super().__init__(message)
        self.status_code = status_code
        self.response_data = response_data or {}


class AuthenticationError(TelecomError):
    pass


class APIError(TelecomError):
    pass


class NetworkError(TelecomError):
    pass


class ValidationError(TelecomError):
    pass


class RateLimitError(TelecomError):
    pass


class ServerError(TelecomError):
    pass
