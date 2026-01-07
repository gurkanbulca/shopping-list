package shopping.platform

/**
 * Interface for secure storage of authentication tokens.
 * Platform-specific implementations handle the actual secure storage mechanism.
 */
interface SecureTokenStore {
    /**
     * Store the access token securely.
     */
    suspend fun storeAccessToken(token: String)

    /**
     * Retrieve the stored access token.
     * @return The token, or null if not stored
     */
    suspend fun getAccessToken(): String?

    /**
     * Store the refresh token securely.
     */
    suspend fun storeRefreshToken(token: String)

    /**
     * Retrieve the stored refresh token.
     * @return The token, or null if not stored
     */
    suspend fun getRefreshToken(): String?

    /**
     * Store the token expiration time.
     * @param expiresAt Unix timestamp in milliseconds
     */
    suspend fun storeExpiresAt(expiresAt: Long)

    /**
     * Get the token expiration time.
     * @return Unix timestamp in milliseconds, or null if not stored
     */
    suspend fun getExpiresAt(): Long?

    /**
     * Clear all stored tokens.
     */
    suspend fun clearAll()

    /**
     * Check if tokens are stored.
     */
    suspend fun hasTokens(): Boolean
}

/**
 * Keys used for token storage.
 */
object TokenStorageKeys {
    const val ACCESS_TOKEN = "auth_access_token"
    const val REFRESH_TOKEN = "auth_refresh_token"
    const val EXPIRES_AT = "auth_expires_at"
}
