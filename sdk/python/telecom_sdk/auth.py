from typing import Optional, Dict
import jwt
from datetime import datetime, timedelta


class AuthProvider:
    """Handles authentication for the Telecom SDK."""
    
    def __init__(
        self,
        api_key: Optional[str] = None,
        jwt_secret: Optional[str] = None,
        jwt_algorithm: str = "HS256",
    ):
        """
        Initialize the authentication provider.
        
        Args:
            api_key: API key for authentication
            jwt_secret: Secret key for JWT token generation/validation
            jwt_algorithm: Algorithm for JWT encoding/decoding
        """
        self.api_key = api_key
        self.jwt_secret = jwt_secret
        self.jwt_algorithm = jwt_algorithm
        self._token_cache: Optional[str] = None
        self._token_expiry: Optional[datetime] = None
    
    def get_headers(self) -> Dict[str, str]:
        """
        Get authentication headers for API requests.
        
        Returns:
            Dictionary of authentication headers
        """
        headers = {
            "Content-Type": "application/json",
            "User-Agent": "Telecom-Python-SDK/1.0.0",
        }
        
        if self.api_key:
            headers["X-API-Key"] = self.api_key
        
        if self._token_cache and self._is_token_valid():
            headers["Authorization"] = f"Bearer {self._token_cache}"
        
        return headers
    
    def generate_jwt_token(
        self,
        user_id: str,
        expiry_hours: int = 24,
        **claims
    ) -> str:
        """
        Generate a JWT token for authentication.
        
        Args:
            user_id: User identifier
            expiry_hours: Token expiry time in hours
            **claims: Additional claims to include in the token
            
        Returns:
            JWT token string
        """
        if not self.jwt_secret:
            raise ValueError("JWT secret not configured")
        
        payload = {
            "sub": user_id,
            "exp": datetime.utcnow() + timedelta(hours=expiry_hours),
            "iat": datetime.utcnow(),
            **claims
        }
        
        token = jwt.encode(payload, self.jwt_secret, algorithm=self.jwt_algorithm)
        self._token_cache = token
        self._token_expiry = datetime.utcnow() + timedelta(hours=expiry_hours)
        
        return token
    
    def validate_jwt_token(self, token: str) -> Dict:
        """
        Validate a JWT token.
        
        Args:
            token: JWT token string
            
        Returns:
            Decoded token payload
            
        Raises:
            ValueError: If token is invalid or expired
        """
        if not self.jwt_secret:
            raise ValueError("JWT secret not configured")
        
        try:
            payload = jwt.decode(
                token,
                self.jwt_secret,
                algorithms=[self.jwt_algorithm]
            )
            return payload
        except jwt.ExpiredSignatureError:
            raise ValueError("Token has expired")
        except jwt.InvalidTokenError as e:
            raise ValueError(f"Invalid token: {e}")
    
    def _is_token_valid(self) -> bool:
        """Check if the cached token is still valid."""
        if not self._token_cache or not self._token_expiry:
            return False
        
        return datetime.utcnow() < self._token_expiry
    
    def clear_token_cache(self) -> None:
        """Clear the cached token."""
        self._token_cache = None
        self._token_expiry = None
