defmodule TelecomSDK.SubscriberAPI do
  @moduledoc """
  API for subscriber management
  """

  defstruct [:client]

  def new(client) do
    %__MODULE__{client: client}
  end

  def get(api, id) do
    TelecomSDK.HTTPClient.get(api.client, "/v1/subscribers/#{id}")
  end

  def list(api, page \\ 1, page_size \\ 50, status \\ nil) do
    params = %{page: page, page_size: page_size}
    params = if status, do: Map.put(params, :status, status), else: params
    TelecomSDK.HTTPClient.get(api.client, "/v1/subscribers", params)
  end

  def create(api, request) do
    TelecomSDK.HTTPClient.post(api.client, "/v1/subscribers", request)
  end

  def update(api, id, request) do
    TelecomSDK.HTTPClient.put(api.client, "/v1/subscribers/#{id}", request)
  end

  def delete(api, id) do
    TelecomSDK.HTTPClient.delete(api.client, "/v1/subscribers/#{id}")
  end

  def suspend(api, id) do
    TelecomSDK.HTTPClient.post(api.client, "/v1/subscribers/#{id}/suspend")
  end

  def activate(api, id) do
    TelecomSDK.HTTPClient.post(api.client, "/v1/subscribers/#{id}/activate")
  end
end
