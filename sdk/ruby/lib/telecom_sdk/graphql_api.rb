module TelecomSDK
  # API for GraphQL queries
  class GraphQLAPI
    def initialize(client)
      @client = client
    end

    def execute(query, variables = nil)
      request = { query: query }
      request[:variables] = variables if variables
      @client.post("/graphql", request)
    end
  end
end
