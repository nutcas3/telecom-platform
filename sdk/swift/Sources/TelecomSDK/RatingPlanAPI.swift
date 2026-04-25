import Foundation

/// API for rating plan management
public class RatingPlanAPI {
    private let client: HTTPClient
    
    public init(client: HTTPClient) {
        self.client = client
    }
    
    /// List all available rating plans
    public func list() async throws -> [RatingPlan] {
        return try await client.get(path: "/v1/rating-plans")
    }
    
    /// Get a rating plan by ID
    public func get(planID: String) async throws -> RatingPlan {
        return try await client.get(path: "/v1/rating-plans/\(planID)")
    }
}
