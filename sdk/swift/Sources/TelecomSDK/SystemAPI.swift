import Foundation

/// API for system management
public class SystemAPI {
    private let client: HTTPClient
    
    public init(client: HTTPClient) {
        self.client = client
    }
    
    /// Get system statistics
    public func getStats() async throws -> SystemStats {
        return try await client.get(path: "/v1/system/stats")
    }
    
    /// Get system health status
    public func getHealth() async throws -> HealthStatus {
        return try await client.get(path: "/v1/health")
    }
}
