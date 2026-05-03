defmodule TelecomSDK.CurrencyAPI do
  @moduledoc """
  Currency and Billing API
  """

  defstruct [:http_client]

  def new(http_client) do
    %__MODULE__{http_client: http_client}
  end

  # Currency Conversion

  def convert(%__MODULE__{http_client: client}, from, to, amount) do
    body = %{from: from, to: to, amount: amount}
    HTTPClient.post(client, "/api/v1/currency/convert", body)
  end

  def get_exchange_rate(%__MODULE__{http_client: client}, from, to) do
    HTTPClient.get(client, "/api/v1/currency/exchange/#{from}/#{to}", %{})
  end

  def get_exchange_rate_history(%__MODULE__{http_client: client}, from, to, days \\ 30) do
    params = %{days: days}
    HTTPClient.get(client, "/api/v1/currency/exchange/#{from}/#{to}/history", params)
  end

  def get_supported_currencies(%__MODULE__{http_client: client}) do
    HTTPClient.get(client, "/api/v1/currency/currencies", %{})
  end

  def refresh_exchange_rates(%__MODULE__{http_client: client}) do
    HTTPClient.post(client, "/api/v1/currency/exchange/refresh", %{})
  end

  # Billing

  def process_billing(%__MODULE__{http_client: client}, billing_data) do
    HTTPClient.post(client, "/api/v1/currency/billing", billing_data)
  end

  def get_billing_history(%__MODULE__{http_client: client}, profile_id, limit \\ 50) do
    params = %{limit: limit}
    HTTPClient.get(client, "/api/v1/currency/billing/history/#{profile_id}", params)
  end

  def get_billing_summary(%__MODULE__{http_client: client}, profile_id, period \\ "monthly") do
    params = %{period: period}
    HTTPClient.get(client, "/api/v1/currency/billing/summary/#{profile_id}", params)
  end

  def process_refund(%__MODULE__{http_client: client}, transaction_id, reason) do
    body = %{reason: reason}
    HTTPClient.post(client, "/api/v1/currency/billing/refund/#{transaction_id}", body)
  end

  def get_billing_analytics(%__MODULE__{http_client: client}, period \\ "monthly") do
    params = %{period: period}
    HTTPClient.get(client, "/api/v1/currency/billing/analytics", params)
  end
end
