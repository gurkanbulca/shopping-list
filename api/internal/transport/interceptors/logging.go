package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs all gRPC requests and responses
func LoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		requestID, _ := ctx.Value(RequestIDKey).(string)
		
		// Log request (mask sensitive data)
		logger.Info("gRPC request started",
			zap.String("method", info.FullMethod),
			zap.String("request_id", requestID),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		// Log response
		if err != nil {
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.String("error", err.Error()),
				zap.String("code", st.Code().String()),
				zap.Duration("duration", duration),
			)
		} else {
			logger.Info("gRPC request completed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}
