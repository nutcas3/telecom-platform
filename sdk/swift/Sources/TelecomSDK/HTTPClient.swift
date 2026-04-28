import Foundation
import NIOHTTP1

/// HTTP client for making API requests
public class HTTPClient {
    private let config: TelecomConfig
    private let authProvider: AuthProvider
    private let session: URLSession
    
    public init(config: TelecomConfig, authProvider: AuthProvider) {
        self.config = config
        self.authProvider = authProvider
        
        let configuration = URLSessionConfiguration.default
        configuration.timeoutIntervalForRequest = config.timeout
        self.session = URLSession(configuration: configuration)
    }
    
    /// Make an HTTP GET request
    public func get<T: Decodable>(path: String, params: [String: String]? = nil) async throws -> T {
        return try await request("GET", path: path, body: nil, params: params)
    }
    
    /// Make an HTTP POST request
    public func post<T: Decodable>(path: String, body: Encodable? = nil) async throws -> T {
        return try await request("POST", path: path, body: body, params: nil)
    }
    
    /// Make an HTTP PUT request
    public func put<T: Decodable>(path: String, body: Encodable? = nil) async throws -> T {
        return try await request("PUT", path: path, body: body, params: nil)
    }
    
    /// Make an HTTP DELETE request
    public func delete(path: String) async throws {
        let _: EmptyResponse = try await request("DELETE", path: path, body: nil, params: nil)
    }
    
    private func request<T: Decodable>(
        _ method: String,
        path: String,
        body: Encodable?,
        params: [String: String]?
    ) async throws -> T {
        var components = URLComponents(string: config.baseURL)
        components?.path = path
        
        if let params = params {
            components?.queryItems = params.map { URLQueryItem(name: $0.key, value: $0.value) }
        }
        
        guard let url = components?.url else {
            throw NetworkError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = method
        
        // Set headers
        let headers = authProvider.getHeaders()
        for (key, value) in headers {
            request.setValue(value, forHTTPHeaderField: key)
        }
        
        // Add body
        if let body = body {
            request.httpBody = try JSONEncoder().encode(body)
        }
        
        // Retry logic
        var lastError: Error?
        for attempt in 0...config.maxRetries {
            do {
                let (data, response) = try await session.data(for: request)
                
                guard let httpResponse = response as? HTTPURLResponse else {
                    throw NetworkError.invalidResponse
                }
                
                try handleResponseErrors(httpResponse)
                
                if T.self == EmptyResponse.self {
                    return EmptyResponse() as! T
                }
                
                return try JSONDecoder().decode(T.self, from: data)
            } catch {
                lastError = error
                if attempt < config.maxRetries {
                    try await Task.sleep(nanoseconds: UInt64(config.retryDelay * 1_000_000_000 * pow(2.0, Double(attempt))))
                }
            }
        }
        
        throw lastError ?? NetworkError.requestFailed
    }
    
    private func handleResponseErrors(_ response: HTTPURLResponse) throws {
        switch response.statusCode {
        case 401:
            throw AuthError.authenticationFailed
        case 429:
            throw AuthError.rateLimitExceeded
        case 400..<500:
            throw NetworkError.clientError(response.statusCode)
        case 500..<600:
            throw NetworkError.serverError(response.statusCode)
        default:
            break
        }
    }
    
    public func close() {
        session.finishTasksAndInvalidate()
    }
}

private struct EmptyResponse: Decodable {}

public enum NetworkError: Error {
    case invalidURL
    case invalidResponse
    case requestFailed
    case clientError(Int)
    case serverError(Int)
}

extension AuthError {
    static let authenticationFailed = AuthError.jwtSecretNotConfigured
    static let rateLimitExceeded = AuthError.tokenExpired
}
