package interceptors

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	tracer       = otel.Tracer("shopping-list-api")
	meter        = otel.Meter("shopping-list-api")
	requestCount metric.Int64Counter
	errorCount   metric.Int64Counter
	latency      metric.Float64Histogram
	activeReqs   metric.Int64UpDownCounter
)

func init() {
	var err error

	// Total request counter
	requestCount, err = meter.Int64Counter(
		"grpc_requests_total",
		metric.WithDescription("Total number of gRPC requests"),
	)
	if err != nil {
		panic(err)
	}

	// Error counter for error rate calculation
	errorCount, err = meter.Int64Counter(
		"grpc_errors_total",
		metric.WithDescription("Total number of gRPC errors"),
	)
	if err != nil {
		panic(err)
	}

	// Latency histogram with buckets suitable for p95/p99 calculation
	// Buckets: 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
	latency, err = meter.Float64Histogram(
		"grpc_request_duration_seconds",
		metric.WithDescription("gRPC request latency in seconds (for p95/p99 calculation)"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}

	// Active requests gauge
	activeReqs, err = meter.Int64UpDownCounter(
		"grpc_active_requests",
		metric.WithDescription("Number of currently active gRPC requests"),
	)
	if err != nil {
		panic(err)
	}
}

// MetricsInterceptor records metrics for all gRPC requests including:
// - Request count (total requests per method)
// - Error rate (errors per method and code)
// - Latency histogram (for p95/p99 calculation)
// - Active requests gauge
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Track active requests
		activeReqs.Add(ctx, 1, metric.WithAttributes(
			attribute.String("method", info.FullMethod),
		))
		defer func() {
			activeReqs.Add(ctx, -1, metric.WithAttributes(
				attribute.String("method", info.FullMethod),
			))
		}()

		// Start span with trace context
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Add request ID to span if available
		if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
			span.SetAttributes(attribute.String("request.id", reqID))
		}

		// Add user ID to span if available
		if userID, ok := GetUserIDFromContext(ctx); ok {
			span.SetAttributes(attribute.String("user.id", userID))
		}

		// Record request
		requestCount.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", info.FullMethod),
			),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		st, _ := status.FromError(err)
		statusCode := st.Code()

		// Record latency with method and status code
		latency.Record(ctx, duration.Seconds(),
			metric.WithAttributes(
				attribute.String("method", info.FullMethod),
				attribute.String("code", statusCode.String()),
			),
		)

		// Record errors for error rate calculation
		if err != nil || statusCode != codes.OK {
			errorCount.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("method", info.FullMethod),
					attribute.String("code", statusCode.String()),
				),
			)
		}

		// Set span attributes
		span.SetAttributes(
			attribute.String("grpc.method", info.FullMethod),
			attribute.String("grpc.code", statusCode.String()),
			attribute.Float64("grpc.duration_seconds", duration.Seconds()),
		)

		if err != nil {
			span.RecordError(err)
		}

		return resp, err
	}
}
