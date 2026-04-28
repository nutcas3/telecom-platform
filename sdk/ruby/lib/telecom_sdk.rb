require "httparty"
require "json"
require "websocket-client-simple"
require "uri"

require_relative "telecom_sdk/auth"
require_relative "telecom_sdk/http_client"
require_relative "telecom_sdk/subscribers_api"
require_relative "telecom_sdk/usage_api"
require_relative "telecom_sdk/payments_api"
require_relative "telecom_sdk/rating_plans_api"
require_relative "telecom_sdk/system_api"
require_relative "telecom_sdk/graphql_api"

module TelecomSDK
  class Error < StandardError; end
  class AuthenticationError < Error; end
  class APIError < Error; end
  class NetworkError < Error; end
  class ValidationError < Error; end
  class RateLimitError < Error; end
  class ServerError < Error; end

  class Client
    attr_reader :config, :subscribers, :usage, :payments, :rating_plans, :system, :graphql

    def initialize(config = {})
      @config = Config.new(config)
      @auth_provider = AuthProvider.new(api_key: @config.api_key, jwt_secret: @config.jwt_secret)
      @http_client = HTTPClient.new(
        @config.api_url,
        @auth_provider,
        timeout: @config.timeout,
        max_retries: @config.max_retries,
        retry_delay: @config.retry_delay,
        enable_logging: @config.enable_logging
      )

      # Initialize API modules
      @subscribers = SubscriberAPI.new(@http_client)
      @usage = UsageAPI.new(@http_client)
      @payments = PaymentAPI.new(@http_client)
      @rating_plans = RatingPlanAPI.new(@http_client)
      @system = SystemAPI.new(@http_client)
      @graphql = GraphQLAPI.new(@http_client)
    end

    # Authentication methods

    def generate_jwt_token(user_id, expiry_hours = 24, additional_claims = {})
      @auth_provider.generate_jwt_token(user_id, expiry_hours, additional_claims)
    end

    def validate_jwt_token(token)
      @auth_provider.validate_jwt_token(token)
    end

    # Cleanup

    def close
      @http_client.close
    end
  end

  class Config
    attr_accessor :api_url, :api_key, :jwt_secret, :timeout, :max_retries, :retry_delay, :enable_logging

    def initialize(config = {})
      @api_url = config[:api_url] || "http://localhost:8000"
      @api_key = config[:api_key]
      @jwt_secret = config[:jwt_secret]
      @timeout = config[:timeout] || 30
      @max_retries = config[:max_retries] || 3
      @retry_delay = config[:retry_delay] || 1
      @enable_logging = config[:enable_logging] || false
    end
  end
end
