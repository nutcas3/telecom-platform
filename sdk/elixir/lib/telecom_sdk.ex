defmodule TelecomSDK do
  @moduledoc """
  Telecom Platform Elixir SDK
  
  Provides GenServer-based async access to the Telecom Platform API with full Elixir pattern matching.
  """

  use GenServer
  require Logger

  # Client API

  @doc """
  Start the Telecom SDK GenServer
  """
  def start_link(opts \\ []) do
    config = struct(Config, opts)
    GenServer.start_link(__MODULE__, config, name: __MODULE__)
  end

  @doc """
  Get subscriber by ID
  """
  def get_subscriber(id) do
    GenServer.call(__MODULE__, {:get_subscriber, id})
  end

  @doc """
  List subscribers with pagination
  """
  def list_subscribers(page \\ 1, page_size \\ 50, status \\ nil) do
    GenServer.call(__MODULE__, {:list_subscribers, page, page_size, status})
  end

  @doc """
  Create a new subscriber
  """
  def create_subscriber(request) do
    GenServer.call(__MODULE__, {:create_subscriber, request})
  end

  @doc """
  Update an existing subscriber
  """
  def update_subscriber(id, request) do
    GenServer.call(__MODULE__, {:update_subscriber, id, request})
  end

  @doc """
  Suspend a subscriber
  """
  def suspend_subscriber(id) do
    GenServer.call(__MODULE__, {:suspend_subscriber, id})
  end

  @doc """
  Activate a suspended subscriber
  """
  def activate_subscriber(id) do
    GenServer.call(__MODULE__, {:activate_subscriber, id})
  end

  @doc """
  Terminate a subscriber
  """
  def terminate_subscriber(id) do
    GenServer.call(__MODULE__, {:terminate_subscriber, id})
  end

  @doc """
  Get usage statistics for a subscriber
  """
  def get_usage_stats(subscriber_id, start_date, end_date) do
    GenServer.call(__MODULE__, {:get_usage_stats, subscriber_id, start_date, end_date})
  end

  @doc """
  List usage events with filtering
  """
  def list_usage_events(opts \\ []) do
    GenServer.call(__MODULE__, {:list_usage_events, opts})
  end

  @doc """
  Get real-time usage for a subscriber
  """
  def get_real_time_usage(subscriber_id) do
    GenServer.call(__MODULE__, {:get_real_time_usage, subscriber_id})
  end

  @doc """
  Create a payment transaction
  """
  def create_payment_transaction(request) do
    GenServer.call(__MODULE__, {:create_payment_transaction, request})
  end

  @doc """
  Get payment transaction by ID
  """
  def get_payment_transaction(transaction_id) do
    GenServer.call(__MODULE__, {:get_payment_transaction, transaction_id})
  end

  @doc """
  List payment transactions with filtering
  """
  def list_payment_transactions(opts \\ []) do
    GenServer.call(__MODULE__, {:list_payment_transactions, opts})
  end

  @doc """
  List all available rating plans
  """
  def list_rating_plans do
    GenServer.call(__MODULE__, :list_rating_plans)
  end

  @doc """
  Get rating plan by ID
  """
  def get_rating_plan(plan_id) do
    GenServer.call(__MODULE__, {:get_rating_plan, plan_id})
  end

  @doc """
  Get system statistics
  """
  def get_system_stats do
    GenServer.call(__MODULE__, :get_system_stats)
  end

  @doc """
  Get system health status
  """
  def get_health_status do
    GenServer.call(__MODULE__, :get_health_status)
  end

  @doc """
  Execute a GraphQL query
  """
  def execute_graphql_query(query, variables \\ %{}) do
    GenServer.call(__MODULE__, {:execute_graphql_query, query, variables})
  end

  @doc """
  Connect to WebSocket for real-time updates
  """
  def connect_websocket(message_handler \\ nil) do
    GenServer.call(__MODULE__, {:connect_websocket, message_handler})
  end

  # GenServer Callbacks

  @impl true
  def init(config) do
    Finch.start_link(name: MyFinch)
    {:ok, %{config: config, finch: MyFinch}}
  end

  @impl true
  def handle_call({:get_subscriber, id}, _from, state) do
    case make_request(:get, "/v1/subscribers/#{id}", nil, state) do
      {:ok, response} -> {:reply, {:ok, Subscriber.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:list_subscribers, page, page_size, status}, _from, state) do
    params = %{page: page, page_size: page_size}
    params = if status, do: Map.put(params, :status, status), else: params
    
    case make_request(:get, "/v1/subscribers", params, state) do
      {:ok, response} -> {:reply, {:ok, SubscriberList.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:create_subscriber, request}, _from, state) do
    case make_request(:post, "/v1/subscribers", request.to_map(), state) do
      {:ok, response} -> {:reply, {:ok, Subscriber.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:update_subscriber, id, request}, _from, state) do
    case make_request(:put, "/v1/subscribers/#{id}", request.to_map(), state) do
      {:ok, response} -> {:reply, {:ok, Subscriber.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:suspend_subscriber, id}, _from, state) do
    case make_request(:post, "/v1/subscribers/#{id}/suspend", nil, state) do
      {:ok, response} -> {:reply, {:ok, Subscriber.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:activate_subscriber, id}, _from, state) do
    case make_request(:post, "/v1/subscribers/#{id}/activate", nil, state) do
      {:ok, response} -> {:reply, {:ok, Subscriber.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:terminate_subscriber, id}, _from, state) do
    case make_request(:delete, "/v1/subscribers/#{id}", nil, state) do
      {:ok, response} -> {:reply, {:ok, response["success"]}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:get_usage_stats, subscriber_id, start_date, end_date}, _from, state) do
    params = %{
      start_date: DateTime.to_iso8601(start_date),
      end_date: DateTime.to_iso8601(end_date)
    }
    
    case make_request(:get, "/v1/subscribers/#{subscriber_id}/usage", params, state) do
      {:ok, response} -> {:reply, {:ok, UsageStats.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:list_usage_events, opts}, _from, state) do
    params = build_params(opts, [:subscriber_id, :usage_type, :start_date, :end_date, :page, :page_size])
    
    case make_request(:get, "/v1/usage/events", params, state) do
      {:ok, response} -> {:reply, {:ok, UsageEventList.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:get_real_time_usage, subscriber_id}, _from, state) do
    case make_request(:get, "/v1/subscribers/#{subscriber_id}/realtime", nil, state) do
      {:ok, response} -> {:reply, {:ok, RealTimeUsage.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:create_payment_transaction, request}, _from, state) do
    case make_request(:post, "/v1/payments/transactions", request.to_map(), state) do
      {:ok, response} -> {:reply, {:ok, PaymentTransaction.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:get_payment_transaction, transaction_id}, _from, state) do
    case make_request(:get, "/v1/payments/transactions/#{transaction_id}", nil, state) do
      {:ok, response} -> {:reply, {:ok, PaymentTransaction.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:list_payment_transactions, opts}, _from, state) do
    params = build_params(opts, [:subscriber_id, :status, :page, :page_size])
    
    case make_request(:get, "/v1/payments/transactions", params, state) do
      {:ok, response} -> {:reply, {:ok, PaymentTransactionList.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call(:list_rating_plans, _from, state) do
    case make_request(:get, "/v1/rating-plans", nil, state) do
      {:ok, response} -> 
        plans = Enum.map(response, &RatingPlan.from_map/1)
        {:reply, {:ok, plans}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:get_rating_plan, plan_id}, _from, state) do
    case make_request(:get, "/v1/rating-plans/#{plan_id}", nil, state) do
      {:ok, response} -> {:reply, {:ok, RatingPlan.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call(:get_system_stats, _from, state) do
    case make_request(:get, "/v1/system/stats", nil, state) do
      {:ok, response} -> {:reply, {:ok, SystemStats.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call(:get_health_status, _from, state) do
    case make_request(:get, "/v1/health", nil, state) do
      {:ok, response} -> {:reply, {:ok, HealthStatus.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:execute_graphql_query, query, variables}, _from, state) do
    request = %{query: query, variables: variables}
    
    case make_request(:post, "/graphql", request, state) do
      {:ok, response} -> {:reply, {:ok, GraphQLResponse.from_map(response)}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  @impl true
  def handle_call({:connect_websocket, message_handler}, _from, state) do
    ws_url = String.replace(state.config.api_url, "http://", "ws://") <> "/ws"
    
    case WebSockex.start_link(ws_url, WebSocketHandler, %{message_handler: message_handler}) do
      {:ok, pid} -> {:reply, {:ok, pid}, state}
      {:error, error} -> {:reply, {:error, error}, state}
    end
  end

  # Private functions

  defp make_request(method, path, params, state) do
    url = state.config.api_url <> path
    
    headers = [
      {"user-agent", "Telecom-Elixir-SDK/1.0.0"},
      {"content-type", "application/json"}
    ]

    headers = if state.config.api_key do
      [{"authorization", "Bearer #{state.config.api_key}"} | headers]
    else
      headers
    end

    request = Finch.build(method, url, headers, Jason.encode!(params || %{}))

    case Finch.request(request, state.finch) do
      {:ok, response} ->
        case response.status do
          200 ->
            body = response.body |> Jason.decode!()
            {:ok, body}
          401 ->
            {:error, :authentication_error}
          429 ->
            {:error, :rate_limit_error}
          status when status in 400..499 ->
            body = response.body |> Jason.decode!()
            {:error, {:api_error, body["error"] || "Bad request"}}
          status when status in 500..599 ->
            {:error, {:server_error, status}}
          _ ->
            {:error, :unknown_error}
        end
      {:error, reason} ->
        {:error, {:network_error, reason}}
    end
  end

  defp build_params(opts, keys) do
    Enum.reduce(keys, %{}, fn key, acc ->
      case Map.get(opts, key) do
        nil -> acc
        value -> Map.put(acc, key, value)
      end
    end)
  end
end
