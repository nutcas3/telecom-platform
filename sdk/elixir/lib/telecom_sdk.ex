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
  Generate a JWT token for authentication
  """
  def generate_jwt_token(user_id, expiry_hours \\ 24, additional_claims \\ %{}) do
    GenServer.call(__MODULE__, {:generate_jwt_token, user_id, expiry_hours, additional_claims})
  end

  @doc """
  Validate a JWT token
  """
  def validate_jwt_token(token) do
    GenServer.call(__MODULE__, {:validate_jwt_token, token})
  end

  @doc """
  Close the SDK and cleanup resources
  """
  def close do
    GenServer.call(__MODULE__, :close)
  end

  # GenServer Callbacks

  @impl true
  def init(config) do
    auth_provider = AuthProvider.new(api_key: config.api_key, jwt_secret: config.jwt_secret)
    http_client = HTTPClient.new(config.api_url, auth_provider,
      timeout: config.timeout,
      max_retries: config.max_retries,
      retry_delay: config.retry_delay,
      enable_logging: config.enable_logging
    )

    # Initialize API modules
    subscribers = SubscriberAPI.new(http_client)
    usage = UsageAPI.new(http_client)
    payments = PaymentAPI.new(http_client)
    rating_plans = RatingPlanAPI.new(http_client)
    system = SystemAPI.new(http_client)
    analytics = AnalyticsAPI.new(http_client)
    security = SecurityAPI.new(http_client)
    currency = CurrencyAPI.new(http_client)

    state = %{
      config: config,
      auth_provider: auth_provider,
      http_client: http_client,
      subscribers: subscribers,
      usage: usage,
      payments: payments,
      rating_plans: rating_plans,
      system: system,
      analytics: analytics,
      security: security,
      currency: currency
    }

    {:ok, state}
  end

  @impl true
  def handle_call({:generate_jwt_token, user_id, expiry_hours, additional_claims}, _from, state) do
    auth_provider = AuthProvider.generate_jwt_token(state.auth_provider, user_id, expiry_hours, additional_claims)
    {:reply, {:ok, auth_provider.token_cache}, %{state | auth_provider: auth_provider}}
  end

  @impl true
  def handle_call({:validate_jwt_token, token}, _from, state) do
    result = AuthProvider.validate_jwt_token(state.auth_provider, token)
    {:reply, {:ok, result}, state}
  end

  @impl true
  def handle_call(:close, _from, state) do
    HTTPClient.close(state.http_client)
    {:reply, :ok, state}
  end

  # Configuration

  defstruct [:api_url, :api_key, :jwt_secret, :timeout, :max_retries, :retry_delay, :enable_logging]

  defmodule Config do
    @moduledoc """
    Configuration for Telecom SDK
    """

    defstruct [
      api_url: "http://localhost:8000",
      api_key: nil,
      jwt_secret: nil,
      timeout: 30_000,
      max_retries: 3,
      retry_delay: 1_000,
      enable_logging: false
    ]
  end
end
