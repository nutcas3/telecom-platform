defmodule TelecomSDK.HTTPClient do
  @moduledoc """
  HTTP client for making API requests
  """

  defstruct [:base_url, :auth_provider, :timeout, :max_retries, :retry_delay, :enable_logging]

  def new(base_url, auth_provider, opts \\ []) do
    %__MODULE__{
      base_url: base_url,
      auth_provider: auth_provider,
      timeout: Keyword.get(opts, :timeout, 30_000),
      max_retries: Keyword.get(opts, :max_retries, 3),
      retry_delay: Keyword.get(opts, :retry_delay, 1_000),
      enable_logging: Keyword.get(opts, :enable_logging, false)
    }
  end

  def get(client, path, params \\ %{}) do
    request(client, "GET", path, nil, params)
  end

  def post(client, path, body \\ nil) do
    request(client, "POST", path, body, nil)
  end

  def put(client, path, body \\ nil) do
    request(client, "PUT", path, body, nil)
  end

  def delete(client, path) do
    request(client, "DELETE", path, nil, nil)
  end

  defp request(client, method, path, body, params) do
    url = client.base_url <> path
    headers = TelecomSDK.AuthProvider.get_headers(client.auth_provider)

    opts = [
      timeout: client.timeout,
      recv_timeout: client.timeout
    ]

    opts = if params != %{}, do: Keyword.put(opts, :params, params), else: opts
    opts = if body, do: Keyword.put(opts, :body, Jason.encode!(body)), else: opts

    last_error = nil

    Enum.reduce_while(0..client.max_retries, nil, fn attempt, _acc ->
      case do_request(method, url, headers, opts) do
        {:ok, response} ->
          case handle_response_errors(response) do
            :ok -> {:halt, {:ok, Jason.decode!(response.body)}}
            {:error, error} -> {:halt, {:error, error}}
          end
        {:error, error} ->
          if attempt < client.max_retries do
            Process.sleep(client.retry_delay * round(:math.pow(2, attempt)))
            {:cont, error}
          else
            {:halt, {:error, error}}
          end
      end
    end)
  end

  defp do_request(method, url, headers, opts) do
    case HTTPoison.request(method, url, Keyword.get(opts, :body), headers, opts) do
      {:ok, response} -> {:ok, response}
      {:error, error} -> {:error, error}
    end
  end

  defp handle_response_errors(%{status_code: status}) when status >= 200 and status < 300, do: :ok
  defp handle_response_errors(%{status_code: 401}), do: {:error, :authentication_failed}
  defp handle_response_errors(%{status_code: 429}), do: {:error, :rate_limit_exceeded}
  defp handle_response_errors(%{status_code: status}) when status >= 400 and status < 500, do: {:error, {:client_error, status}}
  defp handle_response_errors(%{status_code: status}) when status >= 500, do: {:error, {:server_error, status}}

  def close(_client) do
    # HTTPoison doesn't need explicit cleanup
    :ok
  end
end
