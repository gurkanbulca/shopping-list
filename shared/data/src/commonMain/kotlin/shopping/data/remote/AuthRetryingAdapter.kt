package shopping.data.remote

import shopping.domain.model.DomainError
import shopping.platform.Log

/**
 * Wrapper for remote data sources that handles auth token refresh and retry.
 * 
 * Policy:
 * - On 401/Unauthenticated: Attempt token refresh, then retry original call once
 * - On success: Return result
 * - On refresh failure: Propagate AuthError
 * - Bounded retries: Max 1 retry after refresh
 */
class AuthRetryingAdapter<T : Any>(
    private val refreshToken: suspend () -> Result<Unit>,
    private val isAuthError: (Throwable) -> Boolean = { it is AuthenticationException }
) {
    companion object {
        private const val TAG = "AuthRetryingAdapter"
        private const val MAX_RETRIES = 1
    }

    /**
     * Execute a remote call with automatic auth retry.
     */
    suspend fun execute(block: suspend () -> Result<T>): Result<T> {
        var lastError: Throwable? = null
        var retryCount = 0

        while (retryCount <= MAX_RETRIES) {
            val result = block()

            if (result.isSuccess) {
                return result
            }

            val error = result.exceptionOrNull()!!
            lastError = error

            // Check if this is an auth error that might be recoverable
            if (!isAuthError(error)) {
                Log.d(TAG, "Non-auth error, not retrying: ${error.message}")
                return result
            }

            if (retryCount >= MAX_RETRIES) {
                Log.w(TAG, "Max retries reached, returning error")
                return result
            }

            // Attempt token refresh
            Log.d(TAG, "Auth error detected, attempting token refresh")
            val refreshResult = refreshToken()

            if (refreshResult.isFailure) {
                Log.w(TAG, "Token refresh failed", refreshResult.exceptionOrNull())
                return Result.failure(
                    DomainError.AuthError("Token refresh failed: ${refreshResult.exceptionOrNull()?.message}")
                )
            }

            Log.d(TAG, "Token refreshed, retrying original call")
            retryCount++
        }

        return Result.failure(lastError ?: DomainError.UnknownError("Unknown error during auth retry"))
    }
}

/**
 * Exception indicating an authentication error (401/Unauthenticated).
 */
class AuthenticationException(message: String) : Exception(message)

/**
 * Exception indicating a permission error (403/PermissionDenied).
 */
class PermissionDeniedException(val groupId: String, message: String) : Exception(message)

/**
 * Exception indicating a conflict error (409/FailedPrecondition).
 */
class ConflictException(
    val entityType: String,
    val entityId: String,
    message: String
) : Exception(message)

/**
 * Utility functions for checking error types.
 */
object AuthErrors {
    /**
     * Check if an error is an authentication error.
     */
    fun isAuthError(error: Throwable): Boolean {
        return error is AuthenticationException ||
            error.message?.contains("Unauthenticated", ignoreCase = true) == true ||
            error.message?.contains("401", ignoreCase = true) == true
    }

    /**
     * Check if an error is a permission denied error.
     */
    fun isPermissionDenied(error: Throwable): Boolean {
        return error is PermissionDeniedException ||
            error.message?.contains("PermissionDenied", ignoreCase = true) == true ||
            error.message?.contains("403", ignoreCase = true) == true
    }

    /**
     * Check if an error is a conflict/precondition error.
     */
    fun isConflictError(error: Throwable): Boolean {
        return error is ConflictException ||
            error.message?.contains("FailedPrecondition", ignoreCase = true) == true ||
            error.message?.contains("409", ignoreCase = true) == true
    }
}

/**
 * Retry configuration for remote calls.
 */
data class RetryConfig(
    val maxRetries: Int = 3,
    val initialDelayMs: Long = 100,
    val maxDelayMs: Long = 5000,
    val backoffMultiplier: Double = 2.0
)

/**
 * Execute a block with exponential backoff retry for transient errors.
 */
suspend fun <T> withRetry(
    config: RetryConfig = RetryConfig(),
    isRetryable: (Throwable) -> Boolean = { true },
    block: suspend () -> Result<T>
): Result<T> {
    var lastError: Throwable? = null
    var delayMs = config.initialDelayMs

    for (attempt in 0..config.maxRetries) {
        val result = block()

        if (result.isSuccess) {
            return result
        }

        val error = result.exceptionOrNull()!!
        lastError = error

        if (!isRetryable(error) || attempt >= config.maxRetries) {
            return result
        }

        Log.d("Retry", "Attempt ${attempt + 1} failed, retrying in ${delayMs}ms")
        kotlinx.coroutines.delay(delayMs)
        delayMs = (delayMs * config.backoffMultiplier).toLong().coerceAtMost(config.maxDelayMs)
    }

    return Result.failure(lastError ?: Exception("Unknown error"))
}
