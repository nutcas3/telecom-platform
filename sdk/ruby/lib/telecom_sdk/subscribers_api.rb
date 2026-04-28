module TelecomSDK
  # API for subscriber management
  class SubscriberAPI
    def initialize(client)
      @client = client
    end

    def get(id)
      @client.get("/v1/subscribers/#{id}")
    end

    def list(page: 1, page_size: 50, status: nil)
      params = { page: page, page_size: page_size }
      params[:status] = status if status
      @client.get("/v1/subscribers", params)
    end

    def create(request)
      @client.post("/v1/subscribers", request)
    end

    def update(id, request)
      @client.put("/v1/subscribers/#{id}", request)
    end

    def delete(id)
      @client.delete("/v1/subscribers/#{id}")
    end

    def suspend(id)
      @client.post("/v1/subscribers/#{id}/suspend")
    end

    def activate(id)
      @client.post("/v1/subscribers/#{id}/activate")
    end
  end
end
