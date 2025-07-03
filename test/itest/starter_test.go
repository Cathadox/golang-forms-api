package itest

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"salesforge-assignment/internal/api"
	"salesforge-assignment/internal/config"
	"salesforge-assignment/internal/handler"
	"salesforge-assignment/internal/logger"
	"salesforge-assignment/internal/middleware"
	"salesforge-assignment/internal/model"
	"salesforge-assignment/internal/repository"
	"salesforge-assignment/internal/service"
	"testing"
	"time"
)

var testDbConnStr string

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, connStr, err := startPostgresContainer(ctx)
	if err != nil {
		fmt.Printf("Failed to start PostgreSQL container: %v\n", err)
		os.Exit(1)
	}
	testDbConnStr = connStr

	exitCode := m.Run()

	if err := pgContainer.Terminate(ctx); err != nil {
		fmt.Printf("Failed to terminate PostgreSQL container: %v\n", err)
	}

	os.Exit(exitCode)
}

func startPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	pgc, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(2*time.Minute),
		),
	)
	if err != nil {
		return nil, "", err
	}
	connStr, err := pgc.ConnectionString(ctx, "sslmode=disable")
	return pgc, connStr, err
}

type HandlerIntegrationSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
}

func (suite *HandlerIntegrationSuite) SetupSuite() {
	db, err := gorm.Open(pg.Open(testDbConnStr), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	suite.db.Exec("CREATE SCHEMA IF NOT EXISTS authz;")
	err = suite.db.AutoMigrate(&model.CredentialsModel{}, &model.FormModel{}, &model.FormStepModel{})
	suite.Require().NoError(err)

	gin.SetMode(gin.TestMode)
	os.Setenv("JWT_SECRET_KEY", "integration-test-secret")
	disabledLogger := logger.InitLogger("panic", false)
	testConfig := &config.Config{}
	testConfig.Server.PublicUrl = "http://localhost:3000"
	testConfig.Server.BaseURL = "/api/v1"

	credRepo := repository.NewCredentialsRepository(disabledLogger, suite.db)
	seqRepo := repository.NewFormRepository(disabledLogger, suite.db)
	appService := service.NewFormService(disabledLogger, credRepo, seqRepo, testConfig)
	apiHandler := handler.NewFormHandler(appService)

	router := gin.New()
	router.Use(middleware.InjectLogger(disabledLogger))
	router.Use(gin.Recovery())

	baseGroup := router.Group(testConfig.Server.BaseURL)
	api.RegisterHandlers(baseGroup, apiHandler) // Use the generated registration
	suite.router = router
}

func (suite *HandlerIntegrationSuite) TearDownTest() {
	suite.db.Exec("DELETE FROM public.form_steps")
	suite.db.Exec("DELETE FROM public.form")
	suite.db.Exec("DELETE FROM authz.credentials")
}

func TestHandlerIntegration(t *testing.T) {
	suite.Run(t, new(HandlerIntegrationSuite))
}
