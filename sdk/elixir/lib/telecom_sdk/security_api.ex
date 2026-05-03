defmodule TelecomSDK.SecurityAPI do
  @moduledoc """
  Security API for fraud detection and SIM swap protection
  """

  defstruct [:http_client]

  def new(http_client) do
    %__MODULE__{http_client: http_client}
  end

  # Fraud Detection

  def analyze_transaction(%__MODULE__{http_client: client}, transaction) do
    HTTPClient.post(client, "/api/v1/security/fraud/analyze", transaction)
  end

  def get_fraud_alerts(%__MODULE__{http_client: client}, filter \\ nil) do
    body = if filter, do: filter, else: %{}
    HTTPClient.post(client, "/api/v1/security/fraud/alerts", body)
  end

  def update_alert_status(%__MODULE__{http_client: client}, alert_id, status, actions \\ []) do
    body = %{status: status, actions: actions}
    HTTPClient.put(client, "/api/v1/security/fraud/alerts/#{alert_id}", body)
  end

  def get_fraud_metrics(%__MODULE__{http_client: client}, period \\ "monthly") do
    params = %{period: period}
    HTTPClient.get(client, "/api/v1/security/fraud/metrics", params)
  end

  def get_fraud_patterns(%__MODULE__{http_client: client}) do
    HTTPClient.get(client, "/api/v1/security/fraud/patterns", %{})
  end

  # SIM Swap Protection

  def verify_sim_swap(%__MODULE__{http_client: client}, profile_id, msisdn) do
    body = %{profile_id: profile_id, msisdn: msisdn}
    HTTPClient.post(client, "/api/v1/security/simswap/verify", body)
  end

  def get_sim_swap_history(%__MODULE__{http_client: client}, profile_id) do
    HTTPClient.get(client, "/api/v1/security/simswap/history/#{profile_id}", %{})
  end
end
