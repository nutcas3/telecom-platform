module TelecomSDK
  # API for system management
  class SystemAPI
    def initialize(client)
      @client = client
    end

    def get_stats
      @client.get("/v1/system/stats")
    end

    def get_health
      @client.get("/v1/health")
    end
  end
end
