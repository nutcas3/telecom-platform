require "httparty"

module TelecomSDK
  # HTTP client for making API requests
  class HTTPClient
    include HTTParty

    def initialize(base_url, auth_provider, timeout: 30, max_retries: 3, retry_delay: 1, enable_logging: false)
      @base_url = base_url
      @auth_provider = auth_provider
      @timeout = timeout
      @max_retries = max_retries
      @retry_delay = retry_delay
      @enable_logging = enable_logging
    end

    def get(path, params = {})
      request("GET", path, nil, params)
    end

    def post(path, body = nil)
      request("POST", path, body, nil)
    end

    def put(path, body = nil)
      request("PUT", path, body, nil)
    end

    def delete(path)
      request("DELETE", path, nil, nil)
    end

    private

    def request(method, path, body, params)
      url = "#{@base_url}#{path}"
      headers = @auth_provider.get_headers

      options = {
        timeout: @timeout,
        headers: headers
      }

      options[:query] = params if params
      options[:body] = body.to_json if body

      last_error = nil
      (0..@max_retries).each do |attempt|
        begin
          response = HTTParty.send(method.downcase, url, options)
          handle_response_errors(response)
          return response.parsed_response
        rescue StandardError => e
          last_error = e
          if attempt < @max_retries
            sleep(@retry_delay * (2 ** attempt))
          end
        end
      end

      raise last_error || NetworkError, "Request failed after #{@max_retries} retries"
    end

    def handle_response_errors(response)
      case response.code
      when 401
        raise AuthenticationError, "Authentication failed"
      when 429
        raise RateLimitError, "Rate limit exceeded"
      when 400..499
        raise APIError, "API error: #{response.parsed_response}"
      when 500..599
        raise ServerError, "Server error: #{response.code}"
      end
    end

    def close
      # HTTParty doesn't need explicit cleanup
    end
  end
end
