require "httparty"
require "json"
require "websocket-client-simple"
require "uri"

module TelecomSDK
  class Error < StandardError; end
  class AuthenticationError < Error; end
  class APIError < Error; end
  class NetworkError < Error; end
  class ValidationError < Error; end
  class RateLimitError < Error; end
  class ServerError < Error; end

  class Client
    include HTTParty
    
    attr_reader :config
    
    def initialize(config = {})
      @config = Config.new(config)
      self.class.base_uri @config.api_url
      self.class.default_options.update(
        timeout: @config.timeout,
        headers: {
          "User-Agent" => "Telecom-Ruby-SDK/1.0.0",
          "Content-Type" => "application/json"
        }
      )
      
      if @config.api_key
        self.class.default_options[:headers]["Authorization"] = "Bearer #{@config.api_key}"
      end
    end
    
    # Subscriber Management
    
    def get_subscriber(id)
      response = self.class.get("/v1/subscribers/#{id}")
      handle_response(response)
      Subscriber.new(response.parsed_response)
    end
    
    def list_subscribers(page: 1, page_size: 50, status: nil)
      params = { page: page, page_size: page_size }
      params[:status] = status if status
      
      response = self.class.get("/v1/subscribers", query: params)
      handle_response(response)
      SubscriberList.new(response.parsed_response)
    end
    
    def create_subscriber(request)
      response = self.class.post("/v1/subscribers", body: request.to_json)
      handle_response(response)
      Subscriber.new(response.parsed_response)
    end
    
    def update_subscriber(id, request)
      response = self.class.put("/v1/subscribers/#{id}", body: request.to_json)
      handle_response(response)
      Subscriber.new(response.parsed_response)
    end
    
    def suspend_subscriber(id)
      response = self.class.post("/v1/subscribers/#{id}/suspend")
      handle_response(response)
      Subscriber.new(response.parsed_response)
    end
    
    def activate_subscriber(id)
      response = self.class.post("/v1/subscribers/#{id}/activate")
      handle_response(response)
      Subscriber.new(response.parsed_response)
    end
    
    def terminate_subscriber(id)
      response = self.class.delete("/v1/subscribers/#{id}")
      handle_response(response)
      response.parsed_response["success"]
    end
    
    # Usage Management
    
    def get_usage_stats(subscriber_id, start_date, end_date)
      params = {
        start_date: start_date.iso8601,
        end_date: end_date.iso8601
      }
      
      response = self.class.get("/v1/subscribers/#{subscriber_id}/usage", query: params)
      handle_response(response)
      UsageStats.new(response.parsed_response)
    end
    
    def list_usage_events(subscriber_id: nil, usage_type: nil, start_date: nil, end_date: nil, page: 1, page_size: 50)
      params = { page: page, page_size: page_size }
      params[:subscriber_id] = subscriber_id if subscriber_id
      params[:usage_type] = usage_type if usage_type
      params[:start_date] = start_date.iso8601 if start_date
      params[:end_date] = end_date.iso8601 if end_date
      
      response = self.class.get("/v1/usage/events", query: params)
      handle_response(response)
      UsageEventList.new(response.parsed_response)
    end
    
    def get_real_time_usage(subscriber_id)
      response = self.class.get("/v1/subscribers/#{subscriber_id}/realtime")
      handle_response(response)
      RealTimeUsage.new(response.parsed_response)
    end
    
    # Payment Management
    
    def create_payment_transaction(request)
      response = self.class.post("/v1/payments/transactions", body: request.to_json)
      handle_response(response)
      PaymentTransaction.new(response.parsed_response)
    end
    
    def get_payment_transaction(transaction_id)
      response = self.class.get("/v1/payments/transactions/#{transaction_id}")
      handle_response(response)
      PaymentTransaction.new(response.parsed_response)
    end
    
    def list_payment_transactions(subscriber_id: nil, status: nil, page: 1, page_size: 50)
      params = { page: page, page_size: page_size }
      params[:subscriber_id] = subscriber_id if subscriber_id
      params[:status] = status if status
      
      response = self.class.get("/v1/payments/transactions", query: params)
      handle_response(response)
      PaymentTransactionList.new(response.parsed_response)
    end
    
    # Rating Plans
    
    def list_rating_plans
      response = self.class.get("/v1/rating-plans")
      handle_response(response)
      response.parsed_response.map { |plan| RatingPlan.new(plan) }
    end
    
    def get_rating_plan(plan_id)
      response = self.class.get("/v1/rating-plans/#{plan_id}")
      handle_response(response)
      RatingPlan.new(response.parsed_response)
    end
    
    # System Management
    
    def get_system_stats
      response = self.class.get("/v1/system/stats")
      handle_response(response)
      SystemStats.new(response.parsed_response)
    end
    
    def get_health_status
      response = self.class.get("/v1/health")
      handle_response(response)
      HealthStatus.new(response.parsed_response)
    end
    
    # WebSocket Support
    
    def connect_websocket(message_handler: nil)
      ws_url = @config.api_url.gsub("http://", "ws://") + "/ws"
      
      WebSocket::Client::Simple.connect(ws_url) do |ws|
        ws.on :message do |msg|
          begin
            data = JSON.parse(msg.data)
            ws_message = WebSocketMessage.new(data)
            message_handler.call(ws_message) if message_handler
          rescue JSON::ParserError => e
            raise APIError, "Failed to parse WebSocket message: #{e.message}"
          end
        end
      end
    end
    
    # GraphQL Support
    
    def execute_graphql_query(query, variables = {})
      request = { query: query, variables: variables }
      response = self.class.post("/graphql", body: request.to_json)
      handle_response(response)
      GraphQLResponse.new(response.parsed_response)
    end
    
    private
    
    def handle_response(response)
      case response.code
      when 401
        raise AuthenticationError, "Authentication failed"
      when 429
        raise RateLimitError, "Rate limit exceeded"
      when 400..499
        error_data = response.parsed_response["error"] || "Bad request"
        raise APIError, "API error: #{error_data}"
      when 500..599
        raise ServerError, "Server error: #{response.code}"
      end
      
      response
    end
  end
  
  class Config
    attr_accessor :api_url, :api_key, :timeout, :max_retries
    
    def initialize(options = {})
      @api_url = options[:api_url] || "http://localhost:8000"
      @api_key = options[:api_key]
      @timeout = options[:timeout] || 30
      @max_retries = options[:max_retries] || 3
    end
  end
end
