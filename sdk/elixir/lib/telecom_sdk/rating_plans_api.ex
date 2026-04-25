defmodule TelecomSDK.RatingPlanAPI do
  @moduledoc """
  API for rating plan management
  """

  defstruct [:client]

  def new(client) do
    %__MODULE__{client: client}
  end

  def list(api) do
    TelecomSDK.HTTPClient.get(api.client, "/v1/rating-plans")
  end

  def get(api, plan_id) do
    TelecomSDK.HTTPClient.get(api.client, "/v1/rating-plans/#{plan_id}")
  end
end
