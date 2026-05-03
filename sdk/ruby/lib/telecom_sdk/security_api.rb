module TelecomSDK
  class SecurityAPI
    def initialize(http_client)
      @http_client = http_client
    end

    # Fraud Detection
    
    def analyze_transaction(transaction)
      @http_client.post("/api/v1/security/fraud/analyze", transaction)
    end

    def get_fraud_alerts(filter: nil)
      payload = filter&.compact || {}
      @http_client.post("/api/v1/security/fraud/alerts", payload)
    end

    def update_alert_status(alert_id:, status:, actions: [])
      @http_client.put("/api/v1/security/fraud/alerts/#{alert_id}", {
        status: status,
        actions: actions
      })
    end

    def get_fraud_metrics(period: "monthly")
      @http_client.get("/api/v1/security/fraud/metrics", { period: period })
    end

    def get_fraud_patterns
      @http_client.get("/api/v1/security/fraud/patterns")
    end

    # SIM Swap Protection
    
    def verify_sim_swap(profile_id:, msisdn:)
      @http_client.post("/api/v1/security/simswap/verify", {
        profile_id: profile_id,
        msisdn: msisdn
      })
    end

    def get_sim_swap_history(profile_id)
      @http_client.get("/api/v1/security/simswap/history/#{profile_id}")
    end
  end
end
