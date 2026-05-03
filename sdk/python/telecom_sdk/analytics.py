"""Analytics API client for churn prediction, market analysis, and pricing optimization."""

from typing import List, Optional
from datetime import datetime

from .types import (
    ChurnPrediction,
    ChurnMetrics,
    ChurnRiskLevel,
    MarketMetrics,
    PredictiveMaintenanceMetrics,
    PricingOptimizationResult,
    PricingMetrics,
)


class AnalyticsAPI:
    """Analytics API client for churn, market, maintenance, and pricing analytics."""

    def __init__(self, client):
        self._client = client

    def predict_churn(self, profile_id: str) -> ChurnPrediction:
        """Predict churn risk for a profile."""
        response = self._client.post("/api/v1/analytics/churn/predict", {"profile_id": profile_id})
        return ChurnPrediction(**response)

    def get_churn_metrics(self, period: str = "monthly") -> ChurnMetrics:
        """Get churn metrics for a period."""
        response = self._client.get("/api/v1/analytics/churn/metrics", params={"period": period})
        return ChurnMetrics(**response)

    def get_at_risk_customers(
        self, risk_level: ChurnRiskLevel, limit: int = 100
    ) -> List[ChurnPrediction]:
        """Get customers at risk of churning."""
        response = self._client.get(
            "/api/v1/analytics/churn/at-risk",
            params={"risk_level": risk_level.value, "limit": str(limit)},
        )
        return [ChurnPrediction(**item) for item in response]

    def get_market_metrics(self, period: str = "monthly") -> MarketMetrics:
        """Get market penetration metrics."""
        response = self._client.get("/api/v1/analytics/market/metrics", params={"period": period})
        return MarketMetrics(**response)

    def get_competitors(self) -> dict:
        """Get competitor analysis."""
        return self._client.get("/api/v1/analytics/market/competitors")

    def get_market_opportunities(self) -> dict:
        """Get market opportunities."""
        return self._client.get("/api/v1/analytics/market/opportunities")

    def get_maintenance_metrics(self, period: str = "monthly") -> PredictiveMaintenanceMetrics:
        """Get predictive maintenance metrics."""
        response = self._client.get(
            "/api/v1/analytics/maintenance/metrics", params={"period": period}
        )
        return PredictiveMaintenanceMetrics(**response)

    def get_assets_health(self) -> dict:
        """Get assets health status."""
        return self._client.get("/api/v1/analytics/maintenance/assets")

    def get_maintenance_alerts(self) -> dict:
        """Get maintenance alerts."""
        return self._client.get("/api/v1/analytics/maintenance/alerts")

    def predict_failure(self, asset_id: str) -> dict:
        """Predict failure for an asset."""
        return self._client.post(f"/api/v1/analytics/maintenance/predict/{asset_id}", {})

    def get_pricing_metrics(self, period: str = "monthly") -> PricingMetrics:
        """Get pricing optimization metrics."""
        response = self._client.get("/api/v1/analytics/pricing/metrics", params={"period": period})
        return PricingMetrics(**response)

    def optimize_pricing(
        self, rate_plan_ids: List[str], strategy: str = "revenue_maximization"
    ) -> List[PricingOptimizationResult]:
        """Optimize pricing for rate plans."""
        response = self._client.post(
            "/api/v1/analytics/pricing/optimize",
            {"rate_plan_ids": rate_plan_ids, "strategy": strategy},
        )
        return [PricingOptimizationResult(**item) for item in response]

    def get_price_elasticity(self) -> dict:
        """Get price elasticity data."""
        return self._client.get("/api/v1/analytics/pricing/elasticity")
