package main

import (
	"context"
	"fmt"
	"github.com/ebubekir/event-stream/internal/adapter/inbound/http/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/ebubekir/event-stream/cmd/api/docs"
	"github.com/ebubekir/event-stream/internal/adapter/inbound/http/handler"
	chRepo "github.com/ebubekir/event-stream/internal/adapter/outbound/persistence/clickhouse"
	pgRepo "github.com/ebubekir/event-stream/internal/adapter/outbound/persistence/postgres"
	eventApp "github.com/ebubekir/event-stream/internal/application/event"
	eventDomain "github.com/ebubekir/event-stream/internal/domain/event"
	chMigrations "github.com/ebubekir/event-stream/migrations/clickhouse"
	"github.com/ebubekir/event-stream/pkg/clickhouse"
	"github.com/ebubekir/event-stream/pkg/config"
	"github.com/ebubekir/event-stream/pkg/logger"
	"github.com/ebubekir/event-stream/pkg/postgresql"
)

func main() {
	cfg := config.Read()

	// Initialize logger
	if err := logger.Init(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	}); err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Initialize repository and metrics reader based on database type
	var eventRepository eventDomain.EventRepository
	var metricsReader eventDomain.EventMetricsReader

	switch cfg.DatabaseType {
	case config.DatabaseTypePostgres:
		db := postgresql.New(cfg.PostgresSQLUrl, "public")
		if err := db.CheckConnection(); err != nil {
			logger.Fatal("failed to connect to PostgreSQL", zap.Error(err))
		}
		eventRepository = pgRepo.NewEventRepository(db)
		metricsReader = pgRepo.NewMetricsReader(db)
		logger.Info("Using PostgreSQL as event store")

	case config.DatabaseTypeClickhouse:
		db := clickhouse.New(cfg.ClickhouseUrl, "default")
		if err := db.CheckConnection(); err != nil {
			logger.Fatal("failed to connect to ClickHouse", zap.Error(err))
		}

		// Run ClickHouse migrations
		migrator, err := clickhouse.NewMigrator(db, chMigrations.MigrationFS)
		if err != nil {
			logger.Fatal("failed to create migrator", zap.Error(err))
		}
		if err := migrator.Up(context.Background()); err != nil {
			logger.Fatal("failed to run migrations", zap.Error(err))
		}

		eventRepository = chRepo.NewEventRepository(db)
		metricsReader = chRepo.NewMetricsReader(db)
		logger.Info("Using ClickHouse as event store")

	default:
		logger.Fatal("unsupported database type", zap.String("type", string(cfg.DatabaseType)))
	}

	// Swagger settings

	switch cfg.EnvironmentType {
	case config.EnvironmentTypeDev:
		docs.SwaggerInfo.Title = "event-stream [development]"
		docs.SwaggerInfo.Host = "localhost:8080/v1"
		docs.SwaggerInfo.Schemes = []string{"http"}
	case config.EnvironmentTypeProd:
		docs.SwaggerInfo.Title = "event-stream [prod]"
		docs.SwaggerInfo.Host = "localhost:8080/v1"
		docs.SwaggerInfo.Schemes = []string{"https"}
	}

	// Initialize application services
	eventService := eventApp.NewEventService(eventRepository, metricsReader)

	// Initialize HTTP handlers
	eventHandler := handler.NewEventHandler(eventService)

	// Setup Gin router
	api := gin.Default()
	api.Use(middleware.CustomRecovery())
	api.Use(gin.Logger())

	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register routes
	v1 := api.Group("/v1")
	eventHandler.RegisterRoutes(v1)

	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("Starting server", zap.String("address", addr))
	if err := api.Run(addr); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
