module TelecomSDK
  # API for usage management
  class UsageAPI
    def initialize(client)
      @client = client
    end

    def get_stats(subscriber_id, start_date, end_date)
      params = {
        start_date: start_date.iso8601,
        end_date: end_date.iso8601
      }
      @client.get("/v1/subscribers/#{subscriber_id}/usage", params)
    end

    def list_events(subscriber_id: nil, usage_type: nil, start_date: nil, end_date: nil, page: 1, page_size: 50)
      params = { page: page, page_size: page_size }
      params[:subscriber_id] = subscriber_id if subscriber_id
      params[:usage_type] = usage_type if usage_type
      params[:start_date] = start_date.iso8601 if start_date
      params[:end_date] = end_date.iso8601 if end_date
      @client.get("/v1/usage/events", params)
    end

    def get_realtime(subscriber_id)
      @client.get("/v1/subscribers/#{subscriber_id}/realtime")
    end
  end
end
