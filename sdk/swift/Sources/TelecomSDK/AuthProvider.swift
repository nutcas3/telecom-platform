import Foundation
import CryptoKit

/// Authentication provider for the Telecom SDK
public class AuthProvider {
    private var apiKey: String?
    private var jwtSecret: String?
    private var tokenCache: String?
    private var tokenExpiry: Date?
    
    public init(apiKey: String? = nil, jwtSecret: String? = nil) {
        self.apiKey = apiKey
        self.jwtSecret = jwtSecret
    }
    
    /// Get authentication headers for API requests
    public func getHeaders() -> [String: String] {
        var headers: [String: String] = [
            "Content-Type": "application/json",
            "User-Agent": "Telecom-Swift-SDK/1.0.0"
        ]
        
        if let apiKey = apiKey {
            headers["X-API-Key"] = apiKey
        }
        
        if let token = tokenCache, isTokenValid() {
            headers["Authorization"] = "Bearer \(token)"
        }
        
        return headers
    }
    
    /// Generate a JWT token for authentication
    public func generateJWTToken(userID: String, expiryHours: Int, additionalClaims: [String: Any] = [:]) throws -> String {
        guard let jwtSecret = jwtSecret else {
            throw AuthError.jwtSecretNotConfigured
        }
        
        let now = Date()
        let exp = now.addingTimeInterval(TimeInterval(expiryHours * 3600))
        
        var claims: [String: Any] = [
            "sub": userID,
            "exp": Int(exp.timeIntervalSince1970),
            "iat": Int(now.timeIntervalSince1970)
        ]
        
        for (key, value) in additionalClaims {
            claims[key] = value
        }
        
        let header: [String: Any] = [
            "alg": "HS256",
            "typ": "JWT"
        ]
        
        let encodedHeader = base64URLEncode(header)
        let encodedPayload = base64URLEncode(claims)
        let signature = sign("\(encodedHeader).\(encodedPayload)", secret: jwtSecret)
        
        let token = "\(encodedHeader).\(encodedPayload).\(signature)"
        tokenCache = token
        tokenExpiry = exp
        
        return token
    }
    
    /// Validate a JWT token
    public func validateJWTToken(_ token: String) throws -> [String: Any] {
        guard let jwtSecret = jwtSecret else {
            throw AuthError.jwtSecretNotConfigured
        }
        
        let parts = token.split(separator: ".", maxSplits: 2)
        guard parts.count == 3 else {
            throw AuthError.invalidTokenFormat
        }
        
        let encodedHeader = String(parts[0])
        let encodedPayload = String(parts[1])
        let signature = String(parts[2])
        
        let expectedSignature = try sign("\(encodedHeader).\(encodedPayload)", secret: jwtSecret)
        
        guard signature == expectedSignature else {
            throw AuthError.invalidTokenSignature
        }
        
        guard let payloadData = base64URLDecode(encodedPayload) else {
            throw AuthError.failedToDecodePayload
        }
        
        guard let payload = try? JSONSerialization.jsonObject(with: payloadData) as? [String: Any] else {
            throw AuthError.failedToDecodePayload
        }
        
        if let exp = payload["exp"] as? Int {
            if Date().timeIntervalSince1970 > Double(exp) {
                throw AuthError.tokenExpired
            }
        }
        
        return payload
    }
    
    /// Clear the cached JWT token
    public func clearTokenCache() {
        tokenCache = nil
        tokenExpiry = nil
    }
    
    private func isTokenValid() -> Bool {
        guard let tokenCache = tokenCache, let tokenExpiry = tokenExpiry else {
            return false
        }
        return Date() < tokenExpiry
    }
    
    private func base64URLEncode(_ data: Any) -> String {
        guard let jsonData = try? JSONSerialization.data(withJSONObject: data) else {
            return ""
        }
        return jsonData.base64EncodedString()
            .replacingOccurrences(of: "=", with: "")
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
    }
    
    private func base64URLDecode(_ data: String) -> Data? {
        var base64String = data
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        
        while base64String.count % 4 != 0 {
            base64String += "="
        }
        
        return Data(base64Encoded: base64String)
    }
    
    private func sign(_ data: String, secret: String) throws -> String {
        guard let secretData = secret.data(using: .utf8),
              let dataData = data.data(using: .utf8) else {
            throw AuthError.signingFailed
        }
        
        let key = SymmetricKey(data: secretData)
        let signature = HMAC<SHA256>.authenticationCode(for: dataData, using: key)
        
        return Data(signature).base64EncodedString()
            .replacingOccurrences(of: "=", with: "")
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
    }
}

public enum AuthError: Error {
    case jwtSecretNotConfigured
    case invalidTokenFormat
    case invalidTokenSignature
    case failedToDecodePayload
    case tokenExpired
    case signingFailed
}
