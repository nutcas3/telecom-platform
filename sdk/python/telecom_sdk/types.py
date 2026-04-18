from typing import Optional, Dict, Any, List
from datetime import datetime
from pydantic import BaseModel, Field
from enum import Enum

class SubscriberStatus(str, Enum):
    ACTIVE = "active"
    SUSPENDED = "suspended"
    TERMINATED = "terminated"

class UsageType(str, Enum):
    DATA = "data"
    VOICE = "voice"
    SMS = "sms"


class PaymentStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"
    FAILED = "failed"
    REFUNDED = "refunded"


class Subscriber(BaseModel):
    id: int
    imsi: str
    msisdn: str
    first_name: str
    last_name: str
    email: str
    organization_id: Optional[str] = None
    status: SubscriberStatus
    plan_id: int
    balance: float = Field(ge=0.0)
    created_at: datetime
    updated_at: datetime
    
    class Config:
        use_enum_values = True


class UsageStats(BaseModel):
    subscriber_id: str
    data_up: int = Field(ge=0)
    data_down: int = Field(ge=0)
    voice_seconds: int = Field(ge=0)
    sms_count: int = Field(ge=0)
    period_start: datetime
    period_end: datetime
    cost: float = Field(ge=0.0)


class PaymentTransaction(BaseModel):
    id: str
    subscriber_id: str
    amount: float = Field(ge=0.0)
    currency: str = Field(min_length=3, max_length=3)
    status: PaymentStatus
    gateway: str
    transaction_id: Optional[str] = None
    created_at: datetime
    completed_at: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None


class SystemStats(BaseModel):
    active_sessions: int = Field(ge=0)
    total_accounts: int = Field(ge=0)
    blocked_users: int = Field(ge=0)
    low_balance_alerts: int = Field(ge=0)
    uptime: float = Field(ge=0.0)
    cpu_usage: float = Field(ge=0.0, le=100.0)
    memory_usage: float = Field(ge=0.0, le=100.0)


class HealthStatus(BaseModel):
    status: str  # "healthy", "degraded", "unhealthy"
    timestamp: datetime
    checks: Dict[str, Dict[str, Any]]
    uptime: float = Field(ge=0.0)


class WebSocketMessage(BaseModel):
    type: str
    data: Dict[str, Any]
    timestamp: datetime


class UsageEvent(BaseModel):
    id: str
    subscriber_id: str
    usage_type: UsageType
    amount: int = Field(ge=0)
    cost: float = Field(ge=0.0)
    timestamp: datetime
    metadata: Optional[Dict[str, Any]] = None


class RatingPlan(BaseModel):
    plan_id: str
    name: str
    data_rate: float = Field(ge=0.0)
    voice_rate: float = Field(ge=0.0)
    sms_rate: float = Field(ge=0.0)
    monthly_fee: float = Field(ge=0.0)
    data_limit: int = Field(ge=0)
    voice_limit: int = Field(ge=0)
    sms_limit: int = Field(ge=0)


class CurrentSession(BaseModel):
    session_id: str
    start_time: datetime
    data_up: int = Field(ge=0)
    data_down: int = Field(ge=0)
    voice_seconds: int = Field(ge=0)
    sms_count: int = Field(ge=0)


class RealTimeUsage(BaseModel):
    """Real-time usage model."""
    current_session: Optional[CurrentSession] = None
    today_usage: Optional[Dict[str, int]] = None


class APIResponse(BaseModel):
    """Generic API response model."""
    success: bool
    data: Optional[Any] = None
    error: Optional[str] = None
    message: Optional[str] = None


class PaginatedResponse(BaseModel):
    """Paginated response model."""
    items: List[Any]
    total: int = Field(ge=0)
    page: int = Field(ge=1)
    page_size: int = Field(ge=1, le=100)
    has_next: bool
    has_prev: bool
