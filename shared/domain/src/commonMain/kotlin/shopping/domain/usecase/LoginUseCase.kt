package shopping.domain.usecase

import shopping.domain.model.*
import shopping.domain.repo.AuthRepository
import shopping.domain.repo.AuthState

/**
 * Use case for user login.
 * Stores tokens securely and loads user profile.
 */
class LoginUseCase(
    private val authRepository: AuthRepository
) {
    /**
     * Execute login with credentials.
     * @param email User's email address
     * @param password User's password
     * @return Result containing AuthTokens on success
     */
    suspend operator fun invoke(email: String, password: String): Result<AuthTokens> {
        // Validate input
        if (email.isBlank()) {
            return Result.failure(DomainError.ValidationError("Email is required"))
        }
        if (password.isBlank()) {
            return Result.failure(DomainError.ValidationError("Password is required"))
        }
        if (!isValidEmail(email)) {
            return Result.failure(DomainError.ValidationError("Invalid email format"))
        }

        // Perform login
        val credentials = LoginCredentials(email = email, password = password)
        return authRepository.login(credentials)
    }

    /**
     * Check if user is currently authenticated.
     */
    suspend fun isAuthenticated(): Boolean {
        return authRepository.getCurrentUser() != null
    }

    /**
     * Get the current authenticated user.
     */
    suspend fun getCurrentUser(): User? {
        return authRepository.getCurrentUser()
    }

    /**
     * Observe authentication state changes.
     */
    fun observeAuthState() = authRepository.observeAuthState()

    /**
     * Logout the current user.
     */
    suspend fun logout(): Result<Unit> {
        return authRepository.logout()
    }

    private fun isValidEmail(email: String): Boolean {
        // Basic email validation
        return email.contains("@") && email.contains(".")
    }
}
