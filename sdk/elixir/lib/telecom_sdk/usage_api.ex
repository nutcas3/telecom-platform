defmodule TelecomSDK.UsageAPI do
  @moduledoc """
  API for usage management
  """

  defstruct [:client]

  def new(client) do
    %__MODULE__{client: client}
  end

  def get_stats(api, subscriber_id, start_date, end_date) do
    params = %{
      start_date: DateTime.to_iso8601(start_date),
      end_date: DateTime.to_iso8601(end_date)
    }
    TelecomSDK.HTTPClient.get(api.client, "/v1/subscribers/#{subscriber_id}/usage", params)
  end

  def list_events(api, opts \\ []) do
    params = %{
      page: Keyword.get(opts, :page, 1),
      page_size: Keyword.get(opts, :page_size, 50)
    }

    params = if Keyword.get(opts, :subscriber_id), do: Map.put(params, :subscriber_id, Keyword.get(opts, :subscriber_id)), else: params
    params = if Keyword.get(opts, :usage_type), do: Map.put(params, :usage_type, Keyword.get(opts, :usage_type)), else: params
    params = if Keyword.get(opts, :start_date), do: Map.put(params, :start_date, DateTime.to_iso8601(Keyword.get(opts, :start_date))), else: params
    params = if Keyword.get(opts, :end_date), do: Map.put(params, :end_date, DateTime.to_iso8601(Keyword.get(opts, :end_date))), else: params

    TelecomSDK.HTTPClient.get(api.client, "/v1/usage/events", params)
  end

  def get_realtime(api, subscriber_id) do
    TelecomSDK.HTTPClient.get(api.client, "/v1/subscribers/#{subscriber_id}/realtime")
  end
end
