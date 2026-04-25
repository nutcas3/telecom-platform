defmodule TelecomSDK.AuthProvider do
  @moduledoc """
  Authentication provider for the Telecom SDK
  """

  defstruct [:api_key, :jwt_secret, :token_cache, :token_expiry]

  def new(opts \\ []) do
    %__MODULE__{
      api_key: Keyword.get(opts, :api_key),
      jwt_secret: Keyword.get(opts, :jwt_secret),
      token_cache: nil,
      token_expiry: nil
    }
  end

  def get_headers(%__MODULE__{api_key: api_key, token_cache: token_cache, token_expiry: token_expiry}) do
    headers = %{
      "Content-Type" => "application/json",
      "User-Agent" => "Telecom-Elixir-SDK/1.0.0"
    }

    headers = if api_key, do: Map.put(headers, "X-API-Key", api_key), else: headers

    if token_cache && token_valid?(token_expiry) do
      Map.put(headers, "Authorization", "Bearer #{token_cache}")
    else
      headers
    end
  end

  def generate_jwt_token(%__MODULE__{jwt_secret: jwt_secret} = auth, user_id, expiry_hours \\ 24, additional_claims \\ %{}) do
    unless jwt_secret, do: raise(ArgumentError, "JWT secret not configured")

    now = System.system_time(:second)
    exp = now + (expiry_hours * 3600)

    claims = %{
      "sub" => user_id,
      "exp" => exp,
      "iat" => now
    }
    |> Map.merge(additional_claims)

    header = %{"alg" => "HS256", "typ" => "JWT"}

    encoded_header = base64_url_encode(header)
    encoded_payload = base64_url_encode(claims)
    signature = sign("#{encoded_header}.#{encoded_payload}", jwt_secret)

    token = "#{encoded_header}.#{encoded_payload}.#{signature}"

    %{auth | token_cache: token, token_expiry: exp}
  end

  def validate_jwt_token(%__MODULE__{jwt_secret: jwt_secret}, token) do
    unless jwt_secret, do: raise(ArgumentError, "JWT secret not configured")

    parts = String.split(token, ".")
    unless length(parts) == 3, do: raise(ArgumentError, "Invalid token format")

    [encoded_header, encoded_payload, signature] = parts
    expected_signature = sign("#{encoded_header}.#{encoded_payload}", jwt_secret)

    unless signature == expected_signature, do: raise(ArgumentError, "Invalid token signature")

    case base64_url_decode(encoded_payload) do
      {:ok, payload_bytes} ->
        case Jason.decode(payload_bytes) do
          {:ok, payload} ->
            if payload["exp"] && payload["exp"] < System.system_time(:second) do
              raise(ArgumentError, "Token has expired")
            end
            payload
          {:error, _} -> raise(ArgumentError, "Failed to decode payload")
        end
      {:error, _} -> raise(ArgumentError, "Failed to decode payload")
    end
  end

  def clear_token_cache(auth) do
    %{auth | token_cache: nil, token_expiry: nil}
  end

  defp token_valid?(nil), do: false
  defp token_valid?(expiry), do: System.system_time(:second) < expiry

  defp base64_url_encode(data) do
    json = Jason.encode!(data)
    json
    |> Base.url_encode64(padding: false)
  end

  defp base64_url_decode(data) do
    # Add padding if needed
    padded_data = data <> String.duplicate("=", rem(4 - rem(String.length(data), 4), 4))
    Base.url_decode64(padded_data)
  end

  defp sign(data, secret) do
    :crypto.mac(:hmac, :sha256, secret, data)
    |> Base.url_encode64(padding: false)
  end
end
