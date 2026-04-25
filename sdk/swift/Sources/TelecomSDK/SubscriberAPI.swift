import Foundation

/// API for subscriber management
public class SubscriberAPI {
    private let client: HTTPClient
    
    public init(client: HTTPClient) {
        self.client = client
    }
    
    /// Get subscriber by ID
    public func get(id: Int64) async throws -> Subscriber {
        return try await client.get(path: "/v1/subscribers/\(id)")
    }
    
    /// List subscribers with pagination
    public func list(page: Int = 1, pageSize: Int = 50, status: SubscriberStatus? = nil) async throws -> SubscriberList {
        var params: [String: String] = [
            "page": "\(page)",
            "page_size": "\(pageSize)"
        ]
        
        if let status = status {
            params["status"] = "\(status)"
        }
        
        return try await client.get(path: "/v1/subscribers", params: params)
    }
    
    /// Create a new subscriber
    public func create(_ request: CreateSubscriberRequest) async throws -> Subscriber {
        return try await client.post(path: "/v1/subscribers", body: request)
    }
    
    /// Update an existing subscriber
    public func update(id: Int64, request: UpdateSubscriberRequest) async throws -> Subscriber {
        return try await client.put(path: "/v1/subscribers/\(id)", body: request)
    }
    
    /// Delete a subscriber
    public func delete(id: Int64) async throws {
        try await client.delete(path: "/v1/subscribers/\(id)")
    }
    
    /// Suspend a subscriber
    public func suspend(id: Int64) async throws -> Subscriber {
        return try await client.post(path: "/v1/subscribers/\(id)/suspend", body: nil as EmptyRequest?)
    }
    
    /// Activate a suspended subscriber
    public func activate(id: Int64) async throws -> Subscriber {
        return try await client.post(path: "/v1/subscribers/\(id)/activate", body: nil as EmptyRequest?)
    }
}

private struct EmptyRequest: Encodable {}
