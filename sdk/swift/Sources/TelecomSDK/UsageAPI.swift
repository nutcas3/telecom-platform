import Foundation

/// API for usage management
public class UsageAPI {
    private let client: HTTPClient
    
    public init(client: HTTPClient) {
        self.client = client
    }
    
    /// Get usage statistics for a subscriber
    public func getStats(subscriberID: Int64, startDate: Date, endDate: Date) async throws -> UsageStats {
        let formatter = ISO8601DateFormatter()
        let params: [String: String] = [
            "start_date": formatter.string(from: startDate),
            "end_date": formatter.string(from: endDate)
        ]
        
        return try await client.get(path: "/v1/subscribers/\(subscriberID)/usage", params: params)
    }
    
    /// List usage events
    public func listEvents(
        subscriberID: Int64? = nil,
        usageType: UsageType? = nil,
        startDate: Date? = nil,
        endDate: Date? = nil,
        page: Int = 1,
        pageSize: Int = 50
    ) async throws -> UsageEventList {
        var params: [String: String] = [
            "page": "\(page)",
            "page_size": "\(pageSize)"
        ]
        
        if let subscriberID = subscriberID {
            params["subscriber_id"] = "\(subscriberID)"
        }
        if let usageType = usageType {
            params["usage_type"] = "\(usageType)"
        }
        if let startDate = startDate {
            let formatter = ISO8601DateFormatter()
            params["start_date"] = formatter.string(from: startDate)
        }
        if let endDate = endDate {
            let formatter = ISO8601DateFormatter()
            params["end_date"] = formatter.string(from: endDate)
        }
        
        return try await client.get(path: "/v1/usage/events", params: params)
    }
    
    /// Get real-time usage for a subscriber
    public func getRealtime(subscriberID: Int64) async throws -> RealTimeUsage {
        return try await client.get(path: "/v1/subscribers/\(subscriberID)/realtime")
    }
}
