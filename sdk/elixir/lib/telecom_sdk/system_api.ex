defmodule TelecomSDK.SystemAPI do
  @moduledoc """
  API for system management
  """

  defstruct [:client]

  def new(client) do
    %__MODULE__{client: client}
  end

  def get_stats(api) do
    TelecomSDK.HTTPClient.get(api.client, "/v1/system/stats")
  end

  def get_health(api) do
    TelecomSDK.HTTPClient.get(api.client, "/v1/health")
  end
end
