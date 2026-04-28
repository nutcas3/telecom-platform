defmodule TelecomSDK.GraphQLAPI do
  @moduledoc """
  API for GraphQL queries
  """

  defstruct [:client]

  def new(client) do
    %__MODULE__{client: client}
  end

  def execute(api, query, variables \\ nil) do
    request = %{query: query}
    request = if variables, do: Map.put(request, :variables, variables), else: request
    TelecomSDK.HTTPClient.post(api.client, "/graphql", request)
  end
end
