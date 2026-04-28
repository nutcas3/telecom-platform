module TelecomSDK
  # API for payment management
  class PaymentAPI
    def initialize(client)
      @client = client
    end

    def create_transaction(request)
      @client.post("/v1/payments/transactions", request)
    end

    def get_transaction(transaction_id)
      @client.get("/v1/payments/transactions/#{transaction_id}")
    end

    def list_transactions(subscriber_id: nil, status: nil, page: 1, page_size: 50)
      params = { page: page, page_size: page_size }
      params[:subscriber_id] = subscriber_id if subscriber_id
      params[:status] = status if status
      @client.get("/v1/payments/transactions", params)
    end
  end
end
