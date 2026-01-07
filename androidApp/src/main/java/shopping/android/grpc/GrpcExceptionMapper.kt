package shopping.android.grpc

import io.grpc.Status
import io.grpc.StatusException
import io.grpc.StatusRuntimeException
import shopping.domain.model.DomainError
import shopping.domain.model.EntityType
import shopping.domain.model.GroupId

/**
 * Maps gRPC exceptions to domain errors.
 */
object GrpcExceptionMapper {

    /**
     * Map a gRPC exception to a domain error.
     */
    fun mapException(e: Exception): Exception {
        val status = when (e) {
            is StatusException -> e.status
            is StatusRuntimeException -> e.status
            else -> return DomainError.UnknownError(e.message ?: "Unknown error")
        }

        return when (status.code) {
            Status.Code.UNAUTHENTICATED -> 
                DomainError.AuthError(status.description ?: "Authentication required")
            
            Status.Code.PERMISSION_DENIED -> 
                DomainError.PermissionDenied(GroupId("")) // Group ID needs to be extracted from context
            
            Status.Code.NOT_FOUND -> {
                // Try to extract entity info from description
                val description = status.description ?: "Resource not found"
                DomainError.NotFound(EntityType.ITEM, description)
            }
            
            Status.Code.FAILED_PRECONDITION -> 
                DomainError.ConflictError(EntityType.ITEM, status.description ?: "")
            
            Status.Code.INVALID_ARGUMENT -> 
                DomainError.ValidationError(status.description ?: "Invalid argument")
            
            Status.Code.UNAVAILABLE -> 
                DomainError.NetworkError("Service unavailable: ${status.description}")
            
            Status.Code.DEADLINE_EXCEEDED -> 
                DomainError.NetworkError("Request timed out")
            
            Status.Code.INTERNAL -> 
                DomainError.UnknownError("Internal server error: ${status.description}")
            
            else -> 
                DomainError.UnknownError(status.description ?: "Unknown gRPC error: ${status.code}")
        }
    }

    /**
     * Check if an exception indicates an authentication error that should trigger token refresh.
     */
    fun isAuthError(e: Exception): Boolean {
        val status = when (e) {
            is StatusException -> e.status
            is StatusRuntimeException -> e.status
            else -> return false
        }
        return status.code == Status.Code.UNAUTHENTICATED
    }

    /**
     * Check if an exception indicates a permission denied error.
     */
    fun isPermissionDenied(e: Exception): Boolean {
        val status = when (e) {
            is StatusException -> e.status
            is StatusRuntimeException -> e.status
            else -> return false
        }
        return status.code == Status.Code.PERMISSION_DENIED
    }

    /**
     * Check if an exception indicates a conflict (version mismatch).
     */
    fun isConflict(e: Exception): Boolean {
        val status = when (e) {
            is StatusException -> e.status
            is StatusRuntimeException -> e.status
            else -> return false
        }
        return status.code == Status.Code.FAILED_PRECONDITION
    }

    /**
     * Check if an exception is retryable.
     */
    fun isRetryable(e: Exception): Boolean {
        val status = when (e) {
            is StatusException -> e.status
            is StatusRuntimeException -> e.status
            else -> return false
        }
        return status.code in listOf(
            Status.Code.UNAVAILABLE,
            Status.Code.DEADLINE_EXCEEDED,
            Status.Code.RESOURCE_EXHAUSTED
        )
    }
}
