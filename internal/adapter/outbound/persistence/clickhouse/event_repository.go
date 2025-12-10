package clickhouse

import (
	"context"
	"fmt"

	"github.com/ebubekir/event-stream/internal/domain"
	"github.com/ebubekir/event-stream/pkg/clickhouse"
)

// EventRepository implements domain/event.EventRepository for ClickHouse
type EventRepository struct {
	db *clickhouse.ClickHouseDb
}

// NewEventRepository creates a new ClickHouse event repository
func NewEventRepository(db *clickhouse.ClickHouseDb) *EventRepository {
	return &EventRepository{db: db}
}

// eventModel is the database model for events in ClickHouse
// ClickHouse uses arrays and nested types natively
type eventModel struct {
	ID                string `db:"id"`
	Name              string `db:"name"`
	ChannelType       string `db:"channel_type"`
	Timestamp         uint16 `db:"timestamp"`
	PreviousTimestamp uint16 `db:"previous_timestamp"`
	Date              string `db:"date"`
	UserID            string `db:"user_id"`
	UserPseudoID      string `db:"user_pseudo_id"`
	// Event Params as parallel arrays (ClickHouse pattern)
	EventParamKeys          []string  `db:"event_param_keys"`
	EventParamStringValues  []string  `db:"event_param_string_values"`
	EventParamNumberValues  []float64 `db:"event_param_number_values"`
	EventParamBooleanValues []uint8   `db:"event_param_boolean_values"`
	// User Params as parallel arrays
	UserParamKeys          []string  `db:"user_param_keys"`
	UserParamStringValues  []string  `db:"user_param_string_values"`
	UserParamNumberValues  []float64 `db:"user_param_number_values"`
	UserParamBooleanValues []uint8   `db:"user_param_boolean_values"`
	// Device fields flattened
	DeviceCategory               string `db:"device_category"`
	DeviceMobileBrandName        string `db:"device_mobile_brand_name"`
	DeviceMobileModelName        string `db:"device_mobile_model_name"`
	DeviceOperatingSystem        string `db:"device_operating_system"`
	DeviceOperatingSystemVersion string `db:"device_operating_system_version"`
	DeviceLanguage               string `db:"device_language"`
	DeviceBrowserName            string `db:"device_browser_name"`
	DeviceBrowserVersion         string `db:"device_browser_version"`
	DeviceHostname               string `db:"device_hostname"`
	// AppInfo fields flattened
	AppInfoID      string `db:"app_info_id"`
	AppInfoVersion string `db:"app_info_version"`
	// Items as parallel arrays
	ItemIDs           []string  `db:"item_ids"`
	ItemNames         []string  `db:"item_names"`
	ItemBrands        []string  `db:"item_brands"`
	ItemVariants      []string  `db:"item_variants"`
	ItemPricesInUsd   []float64 `db:"item_prices_in_usd"`
	ItemQuantities    []int32   `db:"item_quantities"`
	ItemRevenuesInUsd []float64 `db:"item_revenues_in_usd"`
}

// Save persists a single event to ClickHouse
func (r *EventRepository) Save(ctx context.Context, event *domain.Event) error {
	model := toModel(event)

	query := `
		INSERT INTO events (
			id, name, channel_type, timestamp, previous_timestamp, date,
			user_id, user_pseudo_id,
			event_param_keys, event_param_string_values, event_param_number_values, event_param_boolean_values,
			user_param_keys, user_param_string_values, user_param_number_values, user_param_boolean_values,
			device_category, device_mobile_brand_name, device_mobile_model_name,
			device_operating_system, device_operating_system_version,
			device_language, device_browser_name, device_browser_version, device_hostname,
			app_info_id, app_info_version,
			item_ids, item_names, item_brands, item_variants,
			item_prices_in_usd, item_quantities, item_revenues_in_usd
		) VALUES (
			:id, :name, :channel_type, :timestamp, :previous_timestamp, :date,
			:user_id, :user_pseudo_id,
			:event_param_keys, :event_param_string_values, :event_param_number_values, :event_param_boolean_values,
			:user_param_keys, :user_param_string_values, :user_param_number_values, :user_param_boolean_values,
			:device_category, :device_mobile_brand_name, :device_mobile_model_name,
			:device_operating_system, :device_operating_system_version,
			:device_language, :device_browser_name, :device_browser_version, :device_hostname,
			:app_info_id, :app_info_version,
			:item_ids, :item_names, :item_brands, :item_variants,
			:item_prices_in_usd, :item_quantities, :item_revenues_in_usd
		)
	`

	if err := clickhouse.NamedExec(r.db, query, model); err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// SaveBatch persists multiple events using ClickHouse batch insert
func (r *EventRepository) SaveBatch(ctx context.Context, events []*domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	models := make([]eventModel, len(events))
	for i, event := range events {
		models[i] = *toModel(event)
	}

	query := `
		INSERT INTO events (
			id, name, channel_type, timestamp, previous_timestamp, date,
			user_id, user_pseudo_id,
			event_param_keys, event_param_string_values, event_param_number_values, event_param_boolean_values,
			user_param_keys, user_param_string_values, user_param_number_values, user_param_boolean_values,
			device_category, device_mobile_brand_name, device_mobile_model_name,
			device_operating_system, device_operating_system_version,
			device_language, device_browser_name, device_browser_version, device_hostname,
			app_info_id, app_info_version,
			item_ids, item_names, item_brands, item_variants,
			item_prices_in_usd, item_quantities, item_revenues_in_usd
		) VALUES (
			:id, :name, :channel_type, :timestamp, :previous_timestamp, :date,
			:user_id, :user_pseudo_id,
			:event_param_keys, :event_param_string_values, :event_param_number_values, :event_param_boolean_values,
			:user_param_keys, :user_param_string_values, :user_param_number_values, :user_param_boolean_values,
			:device_category, :device_mobile_brand_name, :device_mobile_model_name,
			:device_operating_system, :device_operating_system_version,
			:device_language, :device_browser_name, :device_browser_version, :device_hostname,
			:app_info_id, :app_info_version,
			:item_ids, :item_names, :item_brands, :item_variants,
			:item_prices_in_usd, :item_quantities, :item_revenues_in_usd
		)
	`

	if err := clickhouse.BatchInsert(r.db, query, models); err != nil {
		return fmt.Errorf("failed to batch insert events: %w", err)
	}

	return nil
}

func toModel(event *domain.Event) *eventModel {
	// Convert EventParams to parallel arrays
	eventParamKeys := make([]string, len(event.EventParams))
	eventParamStringValues := make([]string, len(event.EventParams))
	eventParamNumberValues := make([]float64, len(event.EventParams))
	eventParamBooleanValues := make([]uint8, len(event.EventParams))
	for i, p := range event.EventParams {
		eventParamKeys[i] = p.Key
		eventParamStringValues[i] = p.StringValue
		eventParamNumberValues[i] = p.NumberValue
		if p.BooleanValue {
			eventParamBooleanValues[i] = 1
		}
	}

	// Convert UserParams to parallel arrays
	userParamKeys := make([]string, len(event.UserParams))
	userParamStringValues := make([]string, len(event.UserParams))
	userParamNumberValues := make([]float64, len(event.UserParams))
	userParamBooleanValues := make([]uint8, len(event.UserParams))
	for i, p := range event.UserParams {
		userParamKeys[i] = p.Key
		userParamStringValues[i] = p.StringValue
		userParamNumberValues[i] = p.NumberValue
		if p.BooleanValue {
			userParamBooleanValues[i] = 1
		}
	}

	// Convert Items to parallel arrays
	itemIDs := make([]string, len(event.Items))
	itemNames := make([]string, len(event.Items))
	itemBrands := make([]string, len(event.Items))
	itemVariants := make([]string, len(event.Items))
	itemPricesInUsd := make([]float64, len(event.Items))
	itemQuantities := make([]int32, len(event.Items))
	itemRevenuesInUsd := make([]float64, len(event.Items))
	for i, item := range event.Items {
		itemIDs[i] = item.ID
		itemNames[i] = item.Name
		itemBrands[i] = item.Brand
		itemVariants[i] = item.Variant
		itemPricesInUsd[i] = item.PriceInUsd
		itemQuantities[i] = int32(item.Quantity)
		itemRevenuesInUsd[i] = item.RevenueInUsd
	}

	return &eventModel{
		ID:                           event.ID,
		Name:                         event.Name,
		ChannelType:                  string(event.ChannelType),
		Timestamp:                    event.Timestamp,
		PreviousTimestamp:            event.PreviousTimestamp,
		Date:                         event.Date.Format("2006-01-02 15:04:05"),
		UserID:                       event.UserID,
		UserPseudoID:                 event.UserPseudoID,
		EventParamKeys:               eventParamKeys,
		EventParamStringValues:       eventParamStringValues,
		EventParamNumberValues:       eventParamNumberValues,
		EventParamBooleanValues:      eventParamBooleanValues,
		UserParamKeys:                userParamKeys,
		UserParamStringValues:        userParamStringValues,
		UserParamNumberValues:        userParamNumberValues,
		UserParamBooleanValues:       userParamBooleanValues,
		DeviceCategory:               event.Device.Category,
		DeviceMobileBrandName:        event.Device.MobileBrandName,
		DeviceMobileModelName:        event.Device.MobileModelName,
		DeviceOperatingSystem:        event.Device.OperatingSystem,
		DeviceOperatingSystemVersion: event.Device.OperatingSystemVersion,
		DeviceLanguage:               event.Device.Language,
		DeviceBrowserName:            event.Device.BrowserName,
		DeviceBrowserVersion:         event.Device.BrowserVersion,
		DeviceHostname:               event.Device.Hostname,
		AppInfoID:                    event.AppInfo.ID,
		AppInfoVersion:               event.AppInfo.Version,
		ItemIDs:                      itemIDs,
		ItemNames:                    itemNames,
		ItemBrands:                   itemBrands,
		ItemVariants:                 itemVariants,
		ItemPricesInUsd:              itemPricesInUsd,
		ItemQuantities:               itemQuantities,
		ItemRevenuesInUsd:            itemRevenuesInUsd,
	}
}
