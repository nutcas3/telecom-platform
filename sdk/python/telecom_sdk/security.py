"""Security API client for fraud detection and SIM swap protection."""

from typing import List, Optional, Dict, Any
from datetime import datetime

from .types import (
    FraudAlert,
    FraudMetrics,
    FraudAlertFilter,
    FraudType,
    FraudSeverity,
)


class SecurityAPI:
    """Security API client for fraud detection and protection."""

    def __init__(self, client):
        self._client = client

    def analyze_transaction(self, transaction: Dict[str, Any]) -> Optional[FraudAlert]:
        """Analyze a transaction for fraud."""
        response = self._client.post("/api/v1/security/fraud/analyze", transaction)
        if response and response.get("id"):
            return FraudAlert(**response)
        return None

    def get_fraud_alerts(self, filter: Optional[FraudAlertFilter] = None) -> List[FraudAlert]:
        """Get fraud alerts with optional filtering."""
        payload = {}
        if filter:
            payload = filter.model_dump(exclude_none=True)
        response = self._client.post("/api/v1/security/fraud/alerts", payload)
        return [FraudAlert(**item) for item in response]

    def update_alert_status(
        self, alert_id: str, status: str, actions: Optional[List[str]] = None
    ) -> dict:
        """Update fraud alert status."""
        payload = {"status": status}
        if actions:
            payload["actions"] = actions
        return self._client.put(f"/api/v1/security/fraud/alerts/{alert_id}", payload)

    def get_fraud_metrics(self, period: str = "monthly") -> FraudMetrics:
        """Get fraud detection metrics."""
        response = self._client.get("/api/v1/security/fraud/metrics", params={"period": period})
        return FraudMetrics(**response)

    def get_fraud_patterns(self) -> dict:
        """Get detected fraud patterns."""
        return self._client.get("/api/v1/security/fraud/patterns")

    def verify_sim_swap(self, profile_id: str, msisdn: str) -> dict:
        """Verify SIM swap request."""
        return self._client.post(
            "/api/v1/security/simswap/verify",
            {"profile_id": profile_id, "msisdn": msisdn},
        )

    def get_sim_swap_history(self, profile_id: str) -> dict:
        """Get SIM swap history for a profile."""
        return self._client.get(f"/api/v1/security/simswap/history/{profile_id}")
