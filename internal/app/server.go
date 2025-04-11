package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"order-pick-up-point/api/pb"
	"order-pick-up-point/internal/config"
	grpcController "order-pick-up-point/internal/controller/grpc"
	"order-pick-up-point/internal/controller/grpc/middleware"
	controller "order-pick-up-point/internal/controller/http"
	"order-pick-up-point/internal/errs"
	"order-pick-up-point/internal/metrics"
	grpcServ "order-pick-up-point/internal/service/grpc"
	httpServ "order-pick-up-point/internal/service/http"
	"order-pick-up-point/internal/storage/db"
	"order-pick-up-point/pkg/jwt"
	"order-pick-up-point/pkg/logger"
	"order-pick-up-point/pkg/password"
	"time"
)

type Server struct {
	closer       *Closer
	router       *gin.Engine
	pgPool       *pgxpool.Pool
	config       *config.Config
	httpServer   *http.Server
	grpcServer   *grpc.Server
	metricServer *http.Server
	logger       logger.Logger
}

func NewServer(cfg *config.Config, log logger.Logger) *Server {
	c := NewCloser()

	pgPool, err := cfg.Storage.ConnectionToPostgres(log)
	if err != nil {
		log.Fatalw("connect to postgres",
			"error", err)
	}
	c.Add(func(ctx context.Context) error {
		log.Infow("Closing PostgreSQL pool")
		pgPool.Close()
		return nil
	})

	stopChan := make(chan struct{})
	c.Add(func(ctx context.Context) error {
		log.Infow("Stopping DB metrics collection")
		close(stopChan)
		return nil
	})
	dbInterval := time.Duration(cfg.Metrics.DBQueryInterval) * time.Second
	metrics.MonitorDBConnections(pgPool, dbInterval, stopChan)

	txManager := db.NewTxManager(pgPool, log)

	userRepo := db.NewUserRepository(txManager, log)
	pvzRepo := db.NewPvzRepository(txManager, log)
	receptionRepo := db.NewReceptionRepository(txManager, log)
	productRepo := db.NewProductRepository(txManager, log)

	repo := db.NewRepository(userRepo, pvzRepo, receptionRepo, productRepo)

	tokenService := jwt.NewTokenService(cfg.JWT.SecretKey, cfg.JWT.TokenExpiry)
	passwordHasher := password.NewBCryptHasher(0)

	authService := httpServ.NewAuthService(repo, txManager, tokenService, passwordHasher, log, cfg.Allowed.Roles)
	pvzService := httpServ.NewPvzService(repo, txManager, log, cfg.Allowed.Cities, cfg.Allowed.ProductTypes)

	authController := controller.NewAuthController(authService)
	pvzController := controller.NewPvzController(pvzService)

	router := gin.Default()
	SetupRoutes(router, authController, pvzController, tokenService)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.PublicServer.Endpoint, cfg.PublicServer.Port),
		Handler: router,
	}

	pvzGRPCService := grpcServ.NewPvzService(pvzRepo, log)
	pvzGRPCController := grpcController.NewPvzServer(pvzGRPCService)

	grpcSrv := grpc.NewServer()
	pb.RegisterPVZServiceServer(grpcSrv, pvzGRPCController)

	reflection.Register(grpcSrv)

	metricMux := http.NewServeMux()
	metricMux.Handle("/metrics", metrics.MetricsHandler())
	metricServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Metrics.Endpoint, cfg.Metrics.Port),
		Handler: metricMux,
	}

	return &Server{
		closer:       c,
		router:       router,
		pgPool:       pgPool,
		config:       cfg,
		httpServer:   httpServer,
		grpcServer:   grpcSrv,
		metricServer: metricServer,
		logger:       log,
	}
}

func (s *Server) Run(ctx context.Context) error {
	s.closer.Add(func(ctx context.Context) error {
		s.logger.Infow("Shutting down HTTP server")
		return s.httpServer.Shutdown(ctx)
	})

	s.closer.Add(func(ctx context.Context) error {
		s.logger.Infow("Shutting down Metrics server")
		return s.metricServer.Shutdown(ctx)
	})

	go func() {
		s.logger.Infow("Starting HTTP server",
			"address", s.httpServer.Addr,
		)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatalw("HTTP server error",
				"error", err)
		}
	}()

	go func() {
		s.logger.Infow("Starting Metrics server",
			"address", s.metricServer.Addr,
		)
		if err := s.metricServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatalw("Metrics server error",
				"error", err)
		}
	}()

	// Запуск gRPC-сервера
	if err := s.runGRPC(ctx); err != nil {
		return err
	}

	// Запуск gRPC gateway (HTTP-прокси для gRPC)
	if err := s.runGateway(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Server) runGRPC(ctx context.Context) error {
	grpcPort := s.config.GRPCServer.Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return errs.Wrap(err, errs.ErrInternalCode, fmt.Sprintf("failed to listen on port %d", grpcPort))

	}

	s.closer.Add(func(ctx context.Context) error {
		s.logger.Infow("Shutting down gRPC server")
		s.grpcServer.GracefulStop()
		return nil
	})

	go func() {
		s.logger.Infow("Starting gRPC server",
			"address", lis.Addr().String(),
		)
		if err := s.grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			s.logger.Fatalw("gRPC server error",
				"error", err,
			)
		}
	}()

	return nil
}

func (s *Server) runGateway(ctx context.Context) error {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := fmt.Sprintf("%s:%d", s.config.GRPCServer.Endpoint, s.config.GRPCServer.Port)
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return errs.Wrap(err, errs.ErrInternalCode, "failed to dial gRPC endpoint")

	}

	s.closer.Add(func(ctx context.Context) error {
		s.logger.Infow("Closing gRPC connection")
		return conn.Close()
	})

	if err := pb.RegisterPVZServiceHandler(ctx, mux, conn); err != nil {
		return errs.Wrap(err, errs.ErrInternalCode, "failed to register PVZ service handler")
	}

	corsHandler := middleware.EnableCORS(mux)

	gatewayRouter := gin.New()
	gatewayRouter.Use(gin.Logger(), gin.Recovery())

	gatewayRouter.GET("/grpc/listPvz", gin.WrapH(corsHandler))

	gwAddr := fmt.Sprintf(":%d", s.config.Gateway.Port)
	gwServer := &http.Server{
		Addr:    gwAddr,
		Handler: gatewayRouter,
	}

	s.closer.Add(func(ctx context.Context) error {
		s.logger.Infow("Shutting down gRPC gateway")
		return gwServer.Shutdown(ctx)
	})

	go func() {
		s.logger.Infow("Starting gRPC gateway",
			"address", gwAddr,
		)
		if err := gwServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatalw("gRPC gateway error",
				"error", err,
			)
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.closer.Close(ctx)
}
