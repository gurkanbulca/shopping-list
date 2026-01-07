import Foundation
import Security

/// iOS implementation of secure token storage using Keychain.
class IosSecureTokenStore {
    
    private let serviceName = "shopping.ios.tokens"
    
    // MARK: - Public API
    
    func storeAccessToken(_ token: String) async {
        await store(key: TokenStorageKeys.accessToken, value: token)
    }
    
    func getAccessToken() async -> String? {
        return await retrieve(key: TokenStorageKeys.accessToken)
    }
    
    func storeRefreshToken(_ token: String) async {
        await store(key: TokenStorageKeys.refreshToken, value: token)
    }
    
    func getRefreshToken() async -> String? {
        return await retrieve(key: TokenStorageKeys.refreshToken)
    }
    
    func storeExpiresAt(_ expiresAt: Int64) async {
        await store(key: TokenStorageKeys.expiresAt, value: String(expiresAt))
    }
    
    func getExpiresAt() async -> Int64? {
        guard let value = await retrieve(key: TokenStorageKeys.expiresAt) else {
            return nil
        }
        return Int64(value)
    }
    
    func clearAll() async {
        await delete(key: TokenStorageKeys.accessToken)
        await delete(key: TokenStorageKeys.refreshToken)
        await delete(key: TokenStorageKeys.expiresAt)
    }
    
    func hasTokens() async -> Bool {
        let accessToken = await getAccessToken()
        let refreshToken = await getRefreshToken()
        return accessToken != nil && refreshToken != nil
    }
    
    // MARK: - Keychain Operations
    
    private func store(key: String, value: String) async {
        guard let data = value.data(using: .utf8) else { return }
        
        // Delete existing item first
        await delete(key: key)
        
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: serviceName,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlock
        ]
        
        SecItemAdd(query as CFDictionary, nil)
    }
    
    private func retrieve(key: String) async -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: serviceName,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        
        guard status == errSecSuccess,
              let data = result as? Data,
              let value = String(data: data, encoding: .utf8) else {
            return nil
        }
        
        return value
    }
    
    private func delete(key: String) async {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: serviceName,
            kSecAttrAccount as String: key
        ]
        
        SecItemDelete(query as CFDictionary)
    }
}

// MARK: - Token Storage Keys

enum TokenStorageKeys {
    static let accessToken = "auth_access_token"
    static let refreshToken = "auth_refresh_token"
    static let expiresAt = "auth_expires_at"
}

// MARK: - TokenProvider Conformance

extension IosSecureTokenStore: TokenProvider {
    func getAccessToken() async -> String? {
        return await retrieve(key: TokenStorageKeys.accessToken)
    }
}
