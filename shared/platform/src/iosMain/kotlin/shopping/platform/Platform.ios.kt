package shopping.platform

import platform.Foundation.NSUUID
import platform.Foundation.NSDate
import platform.Foundation.timeIntervalSince1970
import platform.Foundation.NSLog
import kotlinx.coroutines.Dispatchers as KotlinDispatchers
import kotlinx.coroutines.newSingleThreadContext

// ============================================
// UUID Generation
// ============================================

actual fun generateUuid(): String = NSUUID().UUIDString.lowercase()

// ============================================
// Time
// ============================================

actual fun currentTimeMillis(): Long = (NSDate().timeIntervalSince1970 * 1000).toLong()

// ============================================
// Logging
// ============================================

private class IosLogger : Logger {
    override fun debug(tag: String, message: String) {
        NSLog("[DEBUG][$tag] ${maskSensitiveData(message)}")
    }

    override fun info(tag: String, message: String) {
        NSLog("[INFO][$tag] ${maskSensitiveData(message)}")
    }

    override fun warn(tag: String, message: String) {
        NSLog("[WARN][$tag] ${maskSensitiveData(message)}")
    }

    override fun error(tag: String, message: String, throwable: Throwable?) {
        val errorMessage = if (throwable != null) {
            "$message: ${throwable.message}"
        } else {
            message
        }
        NSLog("[ERROR][$tag] ${maskSensitiveData(errorMessage)}")
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

actual fun getLogger(): Logger = IosLogger()

// ============================================
// Dispatchers
// ============================================

actual object Dispatchers {
    actual val main: kotlinx.coroutines.CoroutineDispatcher = KotlinDispatchers.Main
    actual val io: kotlinx.coroutines.CoroutineDispatcher = KotlinDispatchers.Default
    actual val default: kotlinx.coroutines.CoroutineDispatcher = KotlinDispatchers.Default
}
