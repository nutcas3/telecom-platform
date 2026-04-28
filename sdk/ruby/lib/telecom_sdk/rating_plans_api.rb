module TelecomSDK
  # API for rating plan management
  class RatingPlanAPI
    def initialize(client)
      @client = client
    end

    def list
      @client.get("/v1/rating-plans")
    end

    def get(plan_id)
      @client.get("/v1/rating-plans/#{plan_id}")
    end
  end
end
