package shopping.platform

import android.util.Log as AndroidLog
import java.util.UUID
import kotlinx.coroutines.Dispatchers as KotlinDispatchers

// ============================================
// UUID Generation
// ============================================

actual fun generateUuid(): String = UUID.randomUUID().toString()

// ============================================
// Time
// ============================================

actual fun currentTimeMillis(): Long = System.currentTimeMillis()

// ============================================
// Logging
// ============================================

private class AndroidLogger : Logger {
    override fun debug(tag: String, message: String) {
        AndroidLog.d(tag, maskSensitiveData(message))
    }

    override fun info(tag: String, message: String) {
        AndroidLog.i(tag, maskSensitiveData(message))
    }

    override fun warn(tag: String, message: String) {
        AndroidLog.w(tag, maskSensitiveData(message))
    }

    override fun error(tag: String, message: String, throwable: Throwable?) {
        if (throwable != null) {
            AndroidLog.e(tag, maskSensitiveData(message), throwable)
        } else {
            AndroidLog.e(tag, maskSensitiveData(message))
        }
    }

    /**
     * Mask sensitive data like tokens and PII before logging.
     */
    private fun maskSensitiveData(message: String): String {
        return message
            // Mask JWT-like tokens
            .replace(Regex("(Bearer\\s+)[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+")) { 
                "${it.groupValues[1]}[REDACTED]" 
            }
            // Mask access/refresh token values
            .replace(Regex("(access_?[Tt]oken|refresh_?[Tt]oken)([\"':=\\s]+)[^\"'\\s,}]+")) { 
                "${it.groupValues[1]}${it.groupValues[2]}[REDACTED]" 
            }
            // Mask email addresses (basic pattern)
            .replace(Regex("[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}")) { 
                "[EMAIL]" 
            }
    }
}

actual fun getLogger(): Logger = AndroidLogger()

// ============================================
// Dispatchers
// ============================================

actual object Dispatchers {
    actual val main: kotlinx.coroutines.CoroutineDispatcher = KotlinDispatchers.Main
    actual val io: kotlinx.coroutines.CoroutineDispatcher = KotlinDispatchers.IO
    actual val default: kotlinx.coroutines.CoroutineDispatcher = KotlinDispatchers.Default
}
