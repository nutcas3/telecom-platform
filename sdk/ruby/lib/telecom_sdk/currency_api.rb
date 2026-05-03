module TelecomSDK
  class CurrencyAPI
    def initialize(http_client)
      @http_client = http_client
    end

    # Currency Conversion
    
    def convert(from:, to:, amount:)
      @http_client.post("/api/v1/currency/convert", {
        from: from,
        to: to,
        amount: amount
      })
    end

    def get_exchange_rate(from:, to:)
      @http_client.get("/api/v1/currency/exchange/#{from}/#{to}")
    end

    def get_exchange_rate_history(from:, to:, days: 30)
      @http_client.get("/api/v1/currency/exchange/#{from}/#{to}/history", { days: days })
    end

    def get_supported_currencies
      @http_client.get("/api/v1/currency/currencies")
    end

    def refresh_exchange_rates
      @http_client.post("/api/v1/currency/exchange/refresh", {})
    end

    # Billing
    
    def process_billing(billing_data)
      @http_client.post("/api/v1/currency/billing", billing_data)
    end

    def get_billing_history(profile_id:, limit: 50)
      @http_client.get("/api/v1/currency/billing/history/#{profile_id}", { limit: limit })
    end

    def get_billing_summary(profile_id:, period: "monthly")
      @http_client.get("/api/v1/currency/billing/summary/#{profile_id}", { period: period })
    end

    def process_refund(transaction_id:, reason:)
      @http_client.post("/api/v1/currency/billing/refund/#{transaction_id}", { reason: reason })
    end

    def get_billing_analytics(period: "monthly")
      @http_client.get("/api/v1/currency/billing/analytics", { period: period })
    end
  end
end
