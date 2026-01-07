package shopping.android.auth

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import shopping.platform.SecureTokenStore
import shopping.platform.TokenStorageKeys

/**
 * Android implementation of SecureTokenStore using EncryptedSharedPreferences.
 * Uses Android Keystore for key management.
 */
class AndroidSecureTokenStore(
    context: Context
) : SecureTokenStore {

    companion object {
        private const val PREFS_FILE_NAME = "shopping_secure_prefs"
    }

    private val masterKey: MasterKey = MasterKey.Builder(context)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .build()

    private val encryptedPrefs: SharedPreferences = EncryptedSharedPreferences.create(
        context,
        PREFS_FILE_NAME,
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    override suspend fun storeAccessToken(token: String) = withContext(Dispatchers.IO) {
        encryptedPrefs.edit()
            .putString(TokenStorageKeys.ACCESS_TOKEN, token)
            .apply()
    }

    override suspend fun getAccessToken(): String? = withContext(Dispatchers.IO) {
        encryptedPrefs.getString(TokenStorageKeys.ACCESS_TOKEN, null)
    }

    override suspend fun storeRefreshToken(token: String) = withContext(Dispatchers.IO) {
        encryptedPrefs.edit()
            .putString(TokenStorageKeys.REFRESH_TOKEN, token)
            .apply()
    }

    override suspend fun getRefreshToken(): String? = withContext(Dispatchers.IO) {
        encryptedPrefs.getString(TokenStorageKeys.REFRESH_TOKEN, null)
    }

    override suspend fun storeExpiresAt(expiresAt: Long) = withContext(Dispatchers.IO) {
        encryptedPrefs.edit()
            .putLong(TokenStorageKeys.EXPIRES_AT, expiresAt)
            .apply()
    }

    override suspend fun getExpiresAt(): Long? = withContext(Dispatchers.IO) {
        if (encryptedPrefs.contains(TokenStorageKeys.EXPIRES_AT)) {
            encryptedPrefs.getLong(TokenStorageKeys.EXPIRES_AT, 0L)
        } else {
            null
        }
    }

    override suspend fun clearAll() = withContext(Dispatchers.IO) {
        encryptedPrefs.edit()
            .remove(TokenStorageKeys.ACCESS_TOKEN)
            .remove(TokenStorageKeys.REFRESH_TOKEN)
            .remove(TokenStorageKeys.EXPIRES_AT)
            .apply()
    }

    override suspend fun hasTokens(): Boolean = withContext(Dispatchers.IO) {
        encryptedPrefs.contains(TokenStorageKeys.ACCESS_TOKEN) &&
            encryptedPrefs.contains(TokenStorageKeys.REFRESH_TOKEN)
    }
}
