package event

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ebubekir/event-stream/internal/domain"
	eventRepo "github.com/ebubekir/event-stream/internal/domain/event"
)

// EventService handles event-related use cases
type EventService struct {
	repo          eventRepo.EventRepository
	metricsReader eventRepo.EventMetricsReader
}

// NewEventService creates a new EventService with the given repository and metrics reader
func NewEventService(repo eventRepo.EventRepository, metricsReader eventRepo.EventMetricsReader) *EventService {
	return &EventService{
		repo:          repo,
		metricsReader: metricsReader,
	}
}

// CreateEvent handles the creation of a new event
func (s *EventService) CreateEvent(ctx context.Context, cmd *CreateEventCommand) (string, error) {
	// Generate unique ID
	id := uuid.New().String()

	// Convert command to domain entity
	event := cmd.ToEvent(id)

	// Persist via repository (PostgreSQL or ClickHouse - service doesn't know)
	if err := s.repo.Save(ctx, event); err != nil {
		return "", fmt.Errorf("failed to save event: %w", err)
	}

	return id, nil
}

// CreateEvents handles batch creation of events
func (s *EventService) CreateEvents(ctx context.Context, cmds []*CreateEventCommand) ([]string, error) {
	if len(cmds) == 0 {
		return []string{}, nil
	}

	ids := make([]string, len(cmds))
	events := make([]*domain.Event, len(cmds))

	for i, cmd := range cmds {
		id := uuid.New().String()
		ids[i] = id
		events[i] = cmd.ToEvent(id)
	}

	if err := s.repo.SaveBatch(ctx, events); err != nil {
		return nil, fmt.Errorf("failed to save events batch: %w", err)
	}

	return ids, nil
}

// GetMetrics retrieves aggregated metrics for events
func (s *EventService) GetMetrics(ctx context.Context, query *GetMetricsQuery) (*MetricsResultDTO, error) {
	result, err := s.metricsReader.GetMetrics(ctx, query.ToMetricsQuery())
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	return FromMetricsResult(result), nil
}
