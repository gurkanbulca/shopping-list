package shopping.platform

/**
 * Platform-specific abstractions using Kotlin expect/actual pattern.
 * These are implemented separately for Android and iOS.
 */

// ============================================
// UUID Generation
// ============================================

/**
 * Generate a random UUID string in canonical format.
 */
expect fun generateUuid(): String

// ============================================
// Time
// ============================================

/**
 * Get current time in milliseconds since Unix epoch.
 */
expect fun currentTimeMillis(): Long

/**
 * Get current time in seconds since Unix epoch.
 */
fun currentTimeSeconds(): Long = currentTimeMillis() / 1000

// ============================================
// Logging
// ============================================

/**
 * Platform-agnostic logger interface.
 */
interface Logger {
    fun debug(tag: String, message: String)
    fun info(tag: String, message: String)
    fun warn(tag: String, message: String)
    fun error(tag: String, message: String, throwable: Throwable? = null)
}

/**
 * Get the platform-specific logger implementation.
 */
expect fun getLogger(): Logger

/**
 * Convenience logging functions.
 */
object Log {
    private val logger: Logger by lazy { getLogger() }

    fun d(tag: String, message: String) = logger.debug(tag, message)
    fun i(tag: String, message: String) = logger.info(tag, message)
    fun w(tag: String, message: String) = logger.warn(tag, message)
    fun e(tag: String, message: String, throwable: Throwable? = null) = 
        logger.error(tag, message, throwable)
}

// ============================================
// Secure Storage (interface only, impl in platform module)
// ============================================

/**
 * Interface for secure key-value storage.
 * Used for storing sensitive data like auth tokens.
 */
interface SecureStorage {
    /**
     * Store a string value securely.
     */
    suspend fun putString(key: String, value: String)

    /**
     * Retrieve a string value.
     * @return The stored value, or null if not found
     */
    suspend fun getString(key: String): String?

    /**
     * Remove a stored value.
     */
    suspend fun remove(key: String)

    /**
     * Check if a key exists.
     */
    suspend fun contains(key: String): Boolean

    /**
     * Clear all stored values.
     */
    suspend fun clear()
}

// ============================================
// Dispatcher Provider
// ============================================

/**
 * Provides platform-specific coroutine dispatchers.
 */
expect object Dispatchers {
    val main: kotlinx.coroutines.CoroutineDispatcher
    val io: kotlinx.coroutines.CoroutineDispatcher
    val default: kotlinx.coroutines.CoroutineDispatcher
}
