package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/ebubekir/event-stream/internal/domain"
	"github.com/ebubekir/event-stream/pkg/postgresql"
)

// EventRepository implements domain/event.EventRepository for PostgreSQL
type EventRepository struct {
	db *postgresql.PostgresDb
}

// NewEventRepository creates a new PostgreSQL event repository
func NewEventRepository(db *postgresql.PostgresDb) *EventRepository {
	return &EventRepository{db: db}
}

// eventModel is the database model for events
type eventModel struct {
	ID                string `db:"id"`
	Name              string `db:"name"`
	ChannelType       string `db:"channel_type"`
	Timestamp         int64  `db:"timestamp"`
	PreviousTimestamp int64  `db:"previous_timestamp"`
	Date              string `db:"date"`
	EventParams       string `db:"event_params"` // JSON
	UserID            string `db:"user_id"`
	UserPseudoID      string `db:"user_pseudo_id"`
	UserParams        string `db:"user_params"` // JSON
	Device            string `db:"device"`      // JSON
	AppInfo           string `db:"app_info"`    // JSON
	Items             string `db:"items"`       // JSON
}

// Save persists a single event to PostgreSQL
func (r *EventRepository) Save(ctx context.Context, event *domain.Event) error {
	model, err := toModel(event)
	if err != nil {
		return fmt.Errorf("failed to convert event to model: %w", err)
	}

	query := `
		INSERT INTO events (
			id, name, channel_type, timestamp, previous_timestamp, date,
			event_params, user_id, user_pseudo_id, user_params,
			device, app_info, items
		) VALUES (
			:id, :name, :channel_type, :timestamp, :previous_timestamp, :date,
			:event_params, :user_id, :user_pseudo_id, :user_params,
			:device, :app_info, :items
		)
	`

	if err := postgresql.NamedExec(r.db, query, model); err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// SaveBatch persists multiple events in a single transaction
func (r *EventRepository) SaveBatch(ctx context.Context, events []*domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	return postgresql.Transaction(r.db, func(tx *sqlx.Tx) error {
		query := `
			INSERT INTO events (
				id, name, channel_type, timestamp, previous_timestamp, date,
				event_params, user_id, user_pseudo_id, user_params,
				device, app_info, items
			) VALUES (
				:id, :name, :channel_type, :timestamp, :previous_timestamp, :date,
				:event_params, :user_id, :user_pseudo_id, :user_params,
				:device, :app_info, :items
			)
		`

		for _, event := range events {
			model, err := toModel(event)
			if err != nil {
				return fmt.Errorf("failed to convert event to model: %w", err)
			}

			if _, err := tx.NamedExecContext(ctx, query, model); err != nil {
				return fmt.Errorf("failed to insert event in batch: %w", err)
			}
		}

		return nil
	})
}

func toModel(event *domain.Event) (*eventModel, error) {
	eventParams, err := json.Marshal(event.EventParams)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event_params: %w", err)
	}

	userParams, err := json.Marshal(event.UserParams)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user_params: %w", err)
	}

	device, err := json.Marshal(event.Device)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device: %w", err)
	}

	appInfo, err := json.Marshal(event.AppInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal app_info: %w", err)
	}

	items, err := json.Marshal(event.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal items: %w", err)
	}

	return &eventModel{
		ID:                event.ID,
		Name:              event.Name,
		ChannelType:       string(event.ChannelType),
		Timestamp:         event.Timestamp,
		PreviousTimestamp: event.PreviousTimestamp,
		Date:              event.Date,
		EventParams:       string(eventParams),
		UserID:            event.UserID,
		UserPseudoID:      event.UserPseudoID,
		UserParams:        string(userParams),
		Device:            string(device),
		AppInfo:           string(appInfo),
		Items:             string(items),
	}, nil
}
