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

class ChurnRiskLevel(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class ChurnPrediction(BaseModel):
    profile_id: str
    risk_level: ChurnRiskLevel
    risk_score: float = Field(ge=0.0, le=100.0)
    predicted_churn_date: Optional[datetime] = None
    reasons: List[str]
    recommendations: List[str]
    last_updated: datetime


class ChurnMetrics(BaseModel):
    period: str
    total_subscribers: int = Field(ge=0)
    churned_subscribers: int = Field(ge=0)
    churn_rate: float = Field(ge=0.0)
    monthly_churn_rate: float = Field(ge=0.0)
    annual_churn_rate: float = Field(ge=0.0)
    average_tenure_days: float = Field(ge=0.0)
    risk_distribution: Dict[ChurnRiskLevel, int]
    generated_at: datetime


class FraudType(str, Enum):
    ACCOUNT_TAKEOVER = "account_takeover"
    SUBSCRIPTION_FRAUD = "subscription_fraud"
    PAYMENT_FRAUD = "payment_fraud"
    USAGE_ANOMALY = "usage_anomaly"
    SIM_SWAP = "sim_swap"


class FraudSeverity(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class FraudAlert(BaseModel):
    id: str
    type: FraudType
    severity: FraudSeverity
    profile_id: str
    description: str
    risk_score: float = Field(ge=0.0, le=100.0)
    evidence: List[str]
    ip_address: str
    timestamp: datetime
    status: str
    actions_taken: List[str]
    metadata: Dict[str, Any]


class FraudMetrics(BaseModel):
    period: str
    total_alerts: int = Field(ge=0)
    resolved_alerts: int = Field(ge=0)
    false_positives: int = Field(ge=0)
    resolution_rate_pct: float = Field(ge=0.0)
    false_positive_rate_pct: float = Field(ge=0.0)
    by_type: Dict[FraudType, int]
    by_severity: Dict[FraudSeverity, int]
    generated_at: datetime


class FraudAlertFilter(BaseModel):
    type: Optional[FraudType] = None
    severity: Optional[FraudSeverity] = None
    status: Optional[str] = None
    from_date: Optional[datetime] = None
    to_date: Optional[datetime] = None
    limit: Optional[int] = Field(None, ge=1, le=1000)


class MarketMetrics(BaseModel):
    period: str
    total_market_size: int = Field(ge=0)
    our_subscribers: int = Field(ge=0)
    market_share_pct: float = Field(ge=0.0)
    growth_rate_pct: float = Field(ge=0.0)
    by_country: Dict[str, "CountryMetrics"]
    by_carrier: Dict[str, "MarketCarrierMetrics"]
    by_demographic: Dict[str, "DemoMetrics"]
    competitor_analysis: Dict[str, "CompetitorMetrics"]
    market_opportunities: List["MarketOpportunity"]
    generated_at: datetime


class CountryMetrics(BaseModel):
    country: str
    market_size: int = Field(ge=0)
    our_subscribers: int = Field(ge=0)
    market_share_pct: float = Field(ge=0.0)
    growth_rate_pct: float = Field(ge=0.0)
    average_revenue: float = Field(ge=0.0)


class MarketCarrierMetrics(BaseModel):
    carrier_id: str
    carrier_name: str
    subscribers: int = Field(ge=0)
    market_share_pct: float = Field(ge=0.0)
    average_revenue: float = Field(ge=0.0)
    quality_score: float = Field(ge=0.0, le=100.0)


class DemoMetrics(BaseModel):
    segment: str
    subscribers: int = Field(ge=0)
    market_share_pct: float = Field(ge=0.0)
    average_revenue: float = Field(ge=0.0)
    growth_rate_pct: float = Field(ge=0.0)


class CompetitorMetrics(BaseModel):
    name: str
    market_share_pct: float = Field(ge=0.0)
    subscribers: int = Field(ge=0)
    average_price: float = Field(ge=0.0)
    strengths: List[str]
    weaknesses: List[str]


class MarketOpportunity(BaseModel):
    id: str
    type: str
    description: str
    potential_size: int = Field(ge=0)
    confidence: float = Field(ge=0.0, le=100.0)
    required_actions: List[str]


class PredictiveMaintenanceMetrics(BaseModel):
    period: str
    total_assets: int = Field(ge=0)
    healthy_assets: int = Field(ge=0)
    at_risk_assets: int = Field(ge=0)
    critical_assets: int = Field(ge=0)
    overall_health_score: float = Field(ge=0.0, le=100.0)
    by_asset_type: Dict[str, "AssetTypeMetrics"]
    predicted_failures: List["PredictedFailure"]
    maintenance_schedule: List["MaintenanceTask"]
    generated_at: datetime


class AssetTypeMetrics(BaseModel):
    asset_type: str
    total: int = Field(ge=0)
    healthy: int = Field(ge=0)
    at_risk: int = Field(ge=0)
    critical: int = Field(ge=0)
    health_score: float = Field(ge=0.0, le=100.0)


class PredictedFailure(BaseModel):
    asset_id: str
    asset_type: str
    failure_type: str
    predicted_date: datetime
    confidence: float = Field(ge=0.0, le=100.0)
    recommended_actions: List[str]


class MaintenanceTask(BaseModel):
    id: str
    asset_id: str
    task_type: str
    priority: str
    scheduled_date: datetime
    estimated_duration_minutes: int = Field(ge=0)
    description: str
    status: str


class PricingOptimizationResult(BaseModel):
    rate_plan_id: str
    current_price: float = Field(ge=0.0)
    optimal_price: float = Field(ge=0.0)
    strategy: str
    expected_revenue: float = Field(ge=0.0)
    expected_demand: int = Field(ge=0)
    price_change_pct: float
    reasoning: List[str]
    risks: List[str]
    recommendations: List[str]
    confidence: float = Field(ge=0.0, le=100.0)
    generated_at: datetime


class PricingMetrics(BaseModel):
    period: str
    total_rate_plans: int = Field(ge=0)
    optimized_rate_plans: int = Field(ge=0)
    average_price_change_pct: float
    expected_revenue_impact_pct: float
    churn_rate_reduction_pct: float
    price_elasticity: float
    competitive_index: float
    optimization_roi_pct: float
    generated_at: datetime
    has_next: bool
    has_prev: bool
