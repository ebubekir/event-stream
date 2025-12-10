package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/ebubekir/event-stream/internal/adapter/inbound/http/handler"
	chRepo "github.com/ebubekir/event-stream/internal/adapter/outbound/persistence/clickhouse"
	pgRepo "github.com/ebubekir/event-stream/internal/adapter/outbound/persistence/postgres"
	eventApp "github.com/ebubekir/event-stream/internal/application/event"
	eventDomain "github.com/ebubekir/event-stream/internal/domain/event"
	"github.com/ebubekir/event-stream/internal/middleware"
	"github.com/ebubekir/event-stream/pkg/clickhouse"
	"github.com/ebubekir/event-stream/pkg/config"
	"github.com/ebubekir/event-stream/pkg/postgresql"
)

func main() {
	cfg := config.Read()

	// Initialize repository based on database type
	var eventRepository eventDomain.EventRepository
	switch cfg.DatabaseType {
	case config.DatabaseTypePostgres:
		db := postgresql.New(cfg.PostgresSQLUrl, "public")
		if err := db.CheckConnection(); err != nil {
			log.Fatalf("failed to connect to PostgreSQL: %v", err)
		}
		eventRepository = pgRepo.NewEventRepository(db)
		log.Println("Using PostgreSQL as event store")

	case config.DatabaseTypeClickhouse:
		db := clickhouse.New(cfg.ClickhouseUrl, "default")
		if err := db.CheckConnection(); err != nil {
			log.Fatalf("failed to connect to ClickHouse: %v", err)
		}
		eventRepository = chRepo.NewEventRepository(db)
		log.Println("Using ClickHouse as event store")

	default:
		log.Fatalf("unsupported database type: %s", cfg.DatabaseType)
	}

	// Initialize application services
	eventService := eventApp.NewEventService(eventRepository)

	// Initialize HTTP handlers
	eventHandler := handler.NewEventHandler(eventService)

	// Setup Gin router
	api := gin.Default()
	api.Use(middleware.CustomRecovery())
	api.Use(gin.Logger())

	// Register routes
	v1 := api.Group("/api/v1")
	eventHandler.RegisterRoutes(v1)

	// TODO: Graceful shutdown
	// TODO: Logging

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	if err := api.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
