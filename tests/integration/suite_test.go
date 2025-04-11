//go:build integration

package integration

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"net/http/httptest"
	"order-pick-up-point/internal/app"
	"order-pick-up-point/internal/config"
	controller "order-pick-up-point/internal/controller/http"
	httpServ "order-pick-up-point/internal/service/http"
	"order-pick-up-point/internal/storage/db"
	"order-pick-up-point/pkg/jwt"
	"order-pick-up-point/pkg/logger"
	"order-pick-up-point/pkg/password"
	"order-pick-up-point/tests/integration/testutil"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	psqlContainer *testutil.PostgreSQLContainer
	server        *httptest.Server
	grpcClient    grpc.ClientConnInterface
	txManager     db.TxManager
	pool          *pgxpool.Pool
}

func (s *TestSuite) SetupSuite() {
	cfg, err := config.LoadConfig("../../configs/", "../../.env")
	s.Require().NoError(err)

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Duration(cfg.PublicServer.ShutdownTimeout)*time.Second)
	defer ctxCancel()

	psqlContainer, err := testutil.NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)

	s.psqlContainer = psqlContainer

	err = testutil.RunMigrations(psqlContainer.GetDSN(), "../../migrations")
	s.Require().NoError(err)

	poolConfig, err := pgxpool.ParseConfig(psqlContainer.GetDSN())
	s.Require().NoError(err)

	poolConfig.MaxConns = int32(cfg.Storage.Postgres.Pool.MaxConnections)
	poolConfig.MinConns = int32(cfg.Storage.Postgres.Pool.MinConnections)
	poolConfig.MaxConnLifetime = time.Duration(cfg.Storage.Postgres.Pool.MaxLifeTime)
	poolConfig.MaxConnIdleTime = time.Duration(cfg.Storage.Postgres.Pool.MaxIdleTime)
	poolConfig.HealthCheckPeriod = time.Duration(cfg.Storage.Postgres.Pool.HealthCheckPeriod)

	pgPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	s.Require().NoError(err)

	s.pool = pgPool

	log := logger.NewLogger(cfg.Env)
	defer log.Sync()

	txManager := db.NewTxManager(pgPool, log)
	s.txManager = txManager

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
	app.SetupRoutes(router, authController, pvzController, tokenService)

	s.server = httptest.NewServer(router)
}

// Выполняется перед каждым тестом
func (s *TestSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer db.Close()

	// Очищаем все таблицы и сбрасываем идентификаторы
	_, err = db.Exec(`
        TRUNCATE TABLE users, pvz, reception, product RESTART IDENTITY CASCADE;
    `)
	s.Require().NoError(err)
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.psqlContainer.Terminate(ctx))

	s.server.Close()
}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) loadFixtures() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer db.Close()

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures/storage"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}
