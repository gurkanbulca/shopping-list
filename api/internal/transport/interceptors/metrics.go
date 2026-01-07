package interceptors

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	tracer  = otel.Tracer("shopping-list-api")
	meter   = otel.Meter("shopping-list-api")
	counter metric.Int64Counter
	latency metric.Float64Histogram
)

func init() {
	var err error
	counter, err = meter.Int64Counter(
		"grpc_requests_total",
		metric.WithDescription("Total number of gRPC requests"),
	)
	if err != nil {
		panic(err)
	}

	latency, err = meter.Float64Histogram(
		"grpc_request_duration_seconds",
		metric.WithDescription("gRPC request latency in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}
}

// MetricsInterceptor records metrics for all gRPC requests
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Start span
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Record request
		counter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", info.FullMethod),
			),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		// Record latency
		latency.Record(ctx, duration.Seconds(),
			metric.WithAttributes(
				attribute.String("method", info.FullMethod),
				attribute.String("code", st.Code().String()),
			),
		)

		// Set span attributes
		span.SetAttributes(
			attribute.String("grpc.method", info.FullMethod),
			attribute.String("grpc.code", st.Code().String()),
			attribute.Float64("grpc.duration", duration.Seconds()),
		)

		if err != nil {
			span.RecordError(err)
		}

		return resp, err
	}
}
