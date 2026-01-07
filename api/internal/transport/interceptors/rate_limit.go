package interceptors

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu           sync.Mutex
	tokens       map[string]float64
	lastRefill   map[string]time.Time
	rate         float64 // tokens per second
	maxTokens    float64 // bucket size
	cleanupEvery time.Duration
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	// AuthRate is the rate limit for auth endpoints (requests per second)
	AuthRate float64
	// AuthBurst is the burst size for auth endpoints
	AuthBurst float64
	// DefaultRate is the rate limit for other endpoints
	DefaultRate float64
	// DefaultBurst is the burst size for other endpoints
	DefaultBurst float64
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		AuthRate:     5,   // 5 requests per second for auth (stricter)
		AuthBurst:    10,  // Allow burst of 10 for auth
		DefaultRate:  100, // 100 requests per second for other endpoints
		DefaultBurst: 200, // Allow burst of 200
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, maxTokens float64) *RateLimiter {
	rl := &RateLimiter{
		tokens:       make(map[string]float64),
		lastRefill:   make(map[string]time.Time),
		rate:         rate,
		maxTokens:    maxTokens,
		cleanupEvery: 10 * time.Minute,
	}
	go rl.cleanup()
	return rl
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Initialize if first request
	if _, exists := rl.tokens[key]; !exists {
		rl.tokens[key] = rl.maxTokens
		rl.lastRefill[key] = now
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(rl.lastRefill[key]).Seconds()
	rl.tokens[key] += elapsed * rl.rate
	if rl.tokens[key] > rl.maxTokens {
		rl.tokens[key] = rl.maxTokens
	}
	rl.lastRefill[key] = now

	// Check if we have tokens available
	if rl.tokens[key] >= 1 {
		rl.tokens[key]--
		return true
	}

	return false
}

// cleanup removes stale entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupEvery)
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.cleanupEvery)
		for key, lastRefill := range rl.lastRefill {
			if lastRefill.Before(cutoff) {
				delete(rl.tokens, key)
				delete(rl.lastRefill, key)
			}
		}
		rl.mu.Unlock()
	}
}

// authMethods lists methods that should have stricter rate limiting
var authMethods = map[string]bool{
	"/shopping.v1.AuthService/Register":     true,
	"/shopping.v1.AuthService/Login":        true,
	"/shopping.v1.AuthService/RefreshToken": true,
}

// RateLimitInterceptor creates a rate limiting interceptor with configurable limits
func RateLimitInterceptor(config RateLimitConfig) grpc.UnaryServerInterceptor {
	authLimiter := NewRateLimiter(config.AuthRate, config.AuthBurst)
	defaultLimiter := NewRateLimiter(config.DefaultRate, config.DefaultBurst)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract client identifier (use request ID or IP for now)
		key := getClientKey(ctx)

		// Choose appropriate limiter based on method
		var limiter *RateLimiter
		if authMethods[info.FullMethod] {
			limiter = authLimiter
		} else {
			limiter = defaultLimiter
		}

		// Check rate limit
		if !limiter.Allow(key) {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded, please try again later")
		}

		return handler(ctx, req)
	}
}

// getClientKey extracts a client identifier from the context
func getClientKey(ctx context.Context) string {
	// Try to get user ID from context (for authenticated requests)
	if userID, ok := GetUserIDFromContext(ctx); ok && userID != "" {
		return "user:" + userID
	}

	// Try to get request ID as fallback
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		return "req:" + reqID
	}

	// Default key for anonymous requests
	return "anonymous"
}
