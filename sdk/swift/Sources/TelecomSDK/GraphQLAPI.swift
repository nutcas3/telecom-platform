import Foundation

/// API for GraphQL queries
public class GraphQLAPI {
    private let client: HTTPClient
    
    public init(client: HTTPClient) {
        self.client = client
    }
    
    /// Execute a GraphQL query
    public func execute(query: String, variables: [String: Any]? = nil) async throws -> [String: Any] {
        let request = GraphQLRequest(query: query, variables: variables)
        return try await client.post(path: "/graphql", body: request)
    }
}
