package interceptors

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Patterns for detecting sensitive data
var (
	emailPattern    = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)
	phonePattern    = regexp.MustCompile(`\+?[1-9]\d{1,14}`)
	jwtPattern      = regexp.MustCompile(`eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`)
	passwordPattern = regexp.MustCompile(`(?i)(password|passwd|secret|token|key)["\s:=]+["']?[^"'\s,}]+["']?`)
)

// sensitiveFields maps field names that should be masked
var sensitiveFields = map[string]bool{
	"password":       true,
	"password_hash":  true,
	"access_token":   true,
	"refresh_token":  true,
	"token":          true,
	"secret":         true,
	"api_key":        true,
	"authorization":  true,
	"email":          true,
	"phone":          true,
	"phone_number":   true,
}

// MaskPII masks personally identifiable information in a string
func MaskPII(input string) string {
	// Mask emails
	result := emailPattern.ReplaceAllStringFunc(input, func(email string) string {
		parts := strings.Split(email, "@")
		if len(parts) == 2 {
			return maskString(parts[0]) + "@" + parts[1]
		}
		return "***@***.***"
	})

	// Mask phone numbers
	result = phonePattern.ReplaceAllStringFunc(result, func(phone string) string {
		if len(phone) <= 4 {
			return phone
		}
		return phone[:2] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-2:]
	})

	// Mask JWTs
	result = jwtPattern.ReplaceAllString(result, "[MASKED_TOKEN]")

	// Mask password-like fields
	result = passwordPattern.ReplaceAllString(result, "$1: [MASKED]")

	return result
}

// maskString masks a string, keeping first and last character
func maskString(s string) string {
	if len(s) <= 2 {
		return "***"
	}
	return string(s[0]) + strings.Repeat("*", len(s)-2) + string(s[len(s)-1])
}

// sanitizeRequest creates a safe-to-log representation of a request
func sanitizeRequest(req interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	if req == nil {
		return result
	}

	val := reflect.ValueOf(req)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := strings.ToLower(field.Name)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Check if field is sensitive
		if sensitiveFields[fieldName] {
			result[field.Name] = "[MASKED]"
			continue
		}

		// Handle different types
		switch fieldValue.Kind() {
		case reflect.String:
			strVal := fieldValue.String()
			if strVal != "" {
				result[field.Name] = MaskPII(strVal)
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			result[field.Name] = fieldValue.Int()
		case reflect.Bool:
			result[field.Name] = fieldValue.Bool()
		default:
			// For complex types, just indicate presence
			if !fieldValue.IsZero() {
				result[field.Name] = fmt.Sprintf("[%s]", fieldValue.Kind())
			}
		}
	}

	return result
}

// LoggingInterceptor logs all gRPC requests and responses with PII masking
func LoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		requestID, _ := ctx.Value(RequestIDKey).(string)
		userID, _ := GetUserIDFromContext(ctx)
		
		// Log request with sanitized data
		sanitizedReq := sanitizeRequest(req)
		logger.Info("gRPC request started",
			zap.String("method", info.FullMethod),
			zap.String("request_id", requestID),
			zap.String("user_id", userID),
			zap.Any("request", sanitizedReq),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		// Log response
		if err != nil {
			// Mask any PII in error messages
			maskedError := MaskPII(err.Error())
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.String("user_id", userID),
				zap.String("error", maskedError),
				zap.String("code", st.Code().String()),
				zap.Duration("duration", duration),
			)
		} else {
			logger.Info("gRPC request completed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.String("user_id", userID),
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}
