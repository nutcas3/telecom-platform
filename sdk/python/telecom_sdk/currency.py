"""Currency and Billing API client."""

from typing import List, Optional, Dict, Any
from datetime import datetime


class CurrencyAPI:
    """Currency and Billing API client."""

    def __init__(self, client):
        self._client = client

    def convert(self, from_currency: str, to_currency: str, amount: float) -> dict:
        """Convert currency."""
        return self._client.post(
            "/api/v1/currency/convert",
            {"from": from_currency, "to": to_currency, "amount": amount},
        )

    def get_exchange_rate(self, from_currency: str, to_currency: str) -> dict:
        """Get exchange rate between currencies."""
        return self._client.get(f"/api/v1/currency/exchange/{from_currency}/{to_currency}")

    def get_exchange_rate_history(
        self, from_currency: str, to_currency: str, days: int = 30
    ) -> dict:
        """Get exchange rate history."""
        return self._client.get(
            f"/api/v1/currency/exchange/{from_currency}/{to_currency}/history",
            params={"days": str(days)},
        )

    def get_supported_currencies(self) -> dict:
        """Get list of supported currencies."""
        return self._client.get("/api/v1/currency/currencies")

    def refresh_exchange_rates(self) -> dict:
        """Refresh exchange rates from external sources."""
        return self._client.post("/api/v1/currency/exchange/refresh", {})

    def process_billing(self, billing_data: Dict[str, Any]) -> dict:
        """Process billing transaction."""
        return self._client.post("/api/v1/currency/billing", billing_data)

    def get_billing_history(self, profile_id: str, limit: int = 50) -> dict:
        """Get billing history for a profile."""
        return self._client.get(
            f"/api/v1/currency/billing/history/{profile_id}",
            params={"limit": str(limit)},
        )

    def get_billing_summary(self, profile_id: str, period: str = "monthly") -> dict:
        """Get billing summary for a profile."""
        return self._client.get(
            f"/api/v1/currency/billing/summary/{profile_id}",
            params={"period": period},
        )

    def process_refund(self, transaction_id: str, reason: str) -> dict:
        """Process a refund for a transaction."""
        return self._client.post(
            f"/api/v1/currency/billing/refund/{transaction_id}",
            {"reason": reason},
        )

    def get_billing_analytics(self, period: str = "monthly") -> dict:
        """Get billing analytics."""
        return self._client.get("/api/v1/currency/billing/analytics", params={"period": period})
