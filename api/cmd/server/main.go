package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gurkanbulca/shopping-list/api/internal/config"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/auth"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/group"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/list"
	"github.com/gurkanbulca/shopping-list/api/internal/storage/postgres"
	"github.com/gurkanbulca/shopping-list/api/internal/transport/grpc"
	"github.com/gurkanbulca/shopping-list/api/internal/transport/interceptors"
	pkgauth "github.com/gurkanbulca/shopping-list/api/pkg/auth"
	shoppingv1 "github.com/gurkanbulca/shopping-list/api/proto/shopping/v1"
	"go.uber.org/zap"
	googlegrpc "google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Setup logger
	logger, err := config.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to setup logger: %v", err)
	}
	defer logger.Sync()

	// Parse JWT expiry durations
	accessExpiry, err := time.ParseDuration(cfg.JWTAccessExpiry)
	if err != nil {
		logger.Fatal("invalid JWT_ACCESS_EXPIRY", zap.Error(err))
	}
	refreshExpiry, err := time.ParseDuration(cfg.JWTRefreshExpiry)
	if err != nil {
		// Handle "7d" format (days are not supported by time.ParseDuration)
		refreshExpiry = 7 * 24 * time.Hour
	}

	// Create token manager
	tokenManager := pkgauth.NewTokenManager(cfg.JWTSecret, accessExpiry, refreshExpiry)

	// Setup database connection
	connString := postgres.ConnectionString(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	db, err := postgres.NewDB(ctx, connString)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("database connection established")

	// Setup repositories
	userRepo := postgres.NewUserRepository(db)
	groupRepo := postgres.NewGroupRepository(db)
	memberRepo := postgres.NewMemberRepository(db)
	userLookupRepo := postgres.NewUserLookupRepository(db)
	listRepo := postgres.NewListRepository(db)
	itemRepo := postgres.NewItemRepository(db)

	// Setup services
	authService := auth.NewService(userRepo, tokenManager, logger)
	groupService := group.NewService(groupRepo, memberRepo, userLookupRepo, logger)
	listService := list.NewService(listRepo, itemRepo, groupService, logger)

	// Setup gRPC handlers
	authHandler := grpc.NewAuthHandler(authService)
	groupHandler := grpc.NewGroupHandler(groupService)
	listHandler := grpc.NewListHandler(listService)

	// Setup gRPC server with interceptors
	grpcServer := googlegrpc.NewServer(
		googlegrpc.UnaryInterceptor(interceptors.RequestIDInterceptor),
		googlegrpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor(logger),
			interceptors.AuthInterceptorWithTokenManager(tokenManager),
			interceptors.AuthorizationInterceptor(),
			interceptors.IdempotencyInterceptor(),
			interceptors.MetricsInterceptor(),
		),
	)

	// Register services
	shoppingv1.RegisterAuthServiceServer(grpcServer, authHandler)
	shoppingv1.RegisterGroupServiceServer(grpcServer, groupHandler)
	shoppingv1.RegisterListServiceServer(grpcServer, listHandler)
	// TODO: Register other services
	// shoppingv1.RegisterCategoryServiceServer(grpcServer, categoryHandler)
	// shoppingv1.RegisterSyncServiceServer(grpcServer, syncHandler)

	// Start server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	logger.Info("gRPC server starting", zap.String("port", cfg.GRPCPort))

	// Graceful shutdown
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down gracefully")
	grpcServer.GracefulStop()
}
