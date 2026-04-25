defmodule TelecomSDK.PaymentAPI do
  @moduledoc """
  API for payment management
  """

  defstruct [:client]

  def new(client) do
    %__MODULE__{client: client}
  end

  def create_transaction(api, request) do
    TelecomSDK.HTTPClient.post(api.client, "/v1/payments/transactions", request)
  end

  def get_transaction(api, transaction_id) do
    TelecomSDK.HTTPClient.get(api.client, "/v1/payments/transactions/#{transaction_id}")
  end

  def list_transactions(api, opts \\ []) do
    params = %{
      page: Keyword.get(opts, :page, 1),
      page_size: Keyword.get(opts, :page_size, 50)
    }

    params = if Keyword.get(opts, :subscriber_id), do: Map.put(params, :subscriber_id, Keyword.get(opts, :subscriber_id)), else: params
    params = if Keyword.get(opts, :status), do: Map.put(params, :status, Keyword.get(opts, :status)), else: params

    TelecomSDK.HTTPClient.get(api.client, "/v1/payments/transactions", params)
  end
end
