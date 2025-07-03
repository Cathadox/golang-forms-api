package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"salesforge-assignment/internal/config"
	"salesforge-assignment/internal/handler"
	"salesforge-assignment/internal/logger"
	"salesforge-assignment/internal/middleware"
	"salesforge-assignment/internal/middleware/auth"
	"salesforge-assignment/internal/repository"
	"salesforge-assignment/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on OS environment variables.")
	}

	cfg, err := config.LoadConfig("server.cfg.yaml")

	if err != nil {
		panic("FATAL: could not load server configuration")
	}

	log := logger.InitLogger(cfg.Log.Level, cfg.Log.Pretty)

	db := OpenDbConnection(log)

	credentialsRepo := repository.NewCredentialsRepository(log, db)
	formRepo := repository.NewFormRepository(log, db)
	apiService := service.NewFormService(log, credentialsRepo, formRepo, cfg)

	apiHandler := handler.NewFormHandler(apiService)

	r := gin.New()

	r.Use(middleware.InjectLogger(log))
	r.Use(middleware.GinLogger())
	r.Use(gin.Recovery())

	baseGroup := r.Group(cfg.Server.BaseURL)
	{
		// --- Public Routes ---
		baseGroup.POST("/login", apiHandler.LoginUser)

		// --- Protected Routes ---
		// Create a new group for all routes that require a valid JWT.
		protected := baseGroup.Group("/")
		protected.Use(auth.AuthMiddleware())
		{
			protected.POST("/form", apiHandler.CreateForm)
			protected.GET("/form/:formId", func(c *gin.Context) {
				apiHandler.GetFormById(c, c.Param("formId"))
			})
			protected.PATCH("/form/:formId", func(c *gin.Context) {
				apiHandler.UpdateFormById(c, c.Param("formId"))
			})
			protected.DELETE("/form/:formId/steps/:stepId", func(c *gin.Context) {
				apiHandler.DeleteFormStepById(c, c.Param("formId"), c.Param("stepId"))
			})
			protected.GET("/form/:formId/steps/:stepId", func(c *gin.Context) {
				apiHandler.GetFormStepById(c, c.Param("formId"), c.Param("stepId"))
			})
			protected.PATCH("/form/:formId/steps/:stepId", func(c *gin.Context) {
				apiHandler.UpdateFormStepById(c, c.Param("formId"), c.Param("stepId"))
			})
		}
	}

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Info().Msgf("Server starting on %s", serverAddr)
	r.Run(serverAddr)

}

func OpenDbConnection(log *zerolog.Logger) *gorm.DB {
	dbDSN := os.Getenv("DATABASE_DSN")
	log.Info().Msgf("Connecting to database with DSN: %s", dbDSN)

	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to the database")
	}

	log.Info().Msg("Database connection established successfully")

	return db
}
