package event

import (
	"context"

	"github.com/ebubekir/event-stream/internal/domain"
)

// EventRepository defines the contract for event persistence
// This interface lives in domain layer - implementations in adapter/outbound
type EventRepository interface {
	// Save persists a single event
	Save(ctx context.Context, event *domain.Event) error

	// SaveBatch persists multiple events in a single operation
	SaveBatch(ctx context.Context, events []*domain.Event) error
}
