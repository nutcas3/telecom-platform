require "openssl"
require "base64"
require "json"

module TelecomSDK
  # Authentication provider for the Telecom SDK
  class AuthProvider
    attr_reader :api_key, :jwt_secret, :token_cache, :token_expiry

    def initialize(api_key: nil, jwt_secret: nil)
      @api_key = api_key
      @jwt_secret = jwt_secret
      @token_cache = nil
      @token_expiry = nil
    end

    # Get authentication headers for API requests
    def get_headers
      headers = {
        "Content-Type" => "application/json",
        "User-Agent" => "Telecom-Ruby-SDK/1.0.0"
      }

      headers["X-API-Key"] = @api_key if @api_key

      if @token_cache && token_valid?
        headers["Authorization"] = "Bearer #{@token_cache}"
      end

      headers
    end

    # Generate a JWT token for authentication
    def generate_jwt_token(user_id, expiry_hours = 24, additional_claims = {})
      raise AuthenticationError, "JWT secret not configured" unless @jwt_secret

      now = Time.now.to_i
      exp = now + (expiry_hours * 3600)

      claims = {
        "sub" => user_id,
        "exp" => exp,
        "iat" => now
      }.merge(additional_claims)

      header = { "alg" => "HS256", "typ" => "JWT" }

      encoded_header = base64_url_encode(header)
      encoded_payload = base64_url_encode(claims)
      signature = sign("#{encoded_header}.#{encoded_payload}")

      @token_cache = "#{encoded_header}.#{encoded_payload}.#{signature}"
      @token_expiry = exp

      @token_cache
    end

    # Validate a JWT token
    def validate_jwt_token(token)
      raise AuthenticationError, "JWT secret not configured" unless @jwt_secret

      parts = token.split(".")
      raise AuthenticationError, "Invalid token format" if parts.size != 3

      encoded_header, encoded_payload, signature = parts
      expected_signature = sign("#{encoded_header}.#{encoded_payload}")

      raise AuthenticationError, "Invalid token signature" unless signature == expected_signature

      payload_bytes = base64_url_decode(encoded_payload)
      raise AuthenticationError, "Failed to decode payload" unless payload_bytes

      payload = JSON.parse(payload_bytes)

      if payload["exp"] && payload["exp"] < Time.now.to_i
        raise AuthenticationError, "Token has expired"
      end

      payload
    end

    # Clear the cached JWT token
    def clear_token_cache
      @token_cache = nil
      @token_expiry = nil
    end

    private

    def token_valid?
      @token_cache && @token_expiry && Time.now.to_i < @token_expiry
    end

    def base64_url_encode(data)
      json = data.to_json
      Base64.urlsafe_encode64(json).gsub("=", "")
    end

    def base64_url_decode(data)
      # Add padding if needed
      padded_data = data + "=" * (4 - data.length % 4)
      Base64.urlsafe_decode64(padded_data)
    end

    def sign(data)
      hmac = OpenSSL::HMAC.new(@jwt_secret, OpenSSL::Digest::SHA256.new)
      hmac.update(data)
      Base64.urlsafe_encode64(hmac.digest).gsub("=", "")
    end
  end
end
