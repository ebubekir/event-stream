package event

import (
	"github.com/ebubekir/event-stream/internal/domain"
)

// CreateEventCommand represents the data needed to create a new event
type CreateEventCommand struct {
	Name              string
	ChannelType       domain.ChannelType
	Timestamp         int64
	PreviousTimestamp int64
	Date              string
	EventParams       []ParamDTO
	UserID            string
	UserPseudoID      string
	UserParams        []ParamDTO
	Device            DeviceDTO
	AppInfo           AppInfoDTO
	Items             []ItemDTO
}

// ParamDTO represents a parameter in application layer
type ParamDTO struct {
	Key          string
	StringValue  string
	NumberValue  float64
	BooleanValue bool
}

// DeviceDTO represents device information in application layer
type DeviceDTO struct {
	Category               string
	MobileBrandName        string
	MobileModelName        string
	OperatingSystem        string
	OperatingSystemVersion string
	Language               string
	BrowserName            string
	BrowserVersion         string
	Hostname               string
}

// AppInfoDTO represents app information in application layer
type AppInfoDTO struct {
	ID      string
	Version string
}

// ItemDTO represents an item in application layer
type ItemDTO struct {
	ID            string
	Name          string
	Brand         string
	Variant       string
	PriceInUsd    float64
	Quantity      int
	RevenueInUsd  float64
	LocationId    string
	ListId        string
	ListName      string
	PromotionId   string
	PromotionName string
	Params        []ParamDTO
}

// ToEvent converts CreateEventCommand to domain.Event
func (c *CreateEventCommand) ToEvent(id string) *domain.Event {
	return &domain.Event{
		ID:                id,
		Name:              c.Name,
		ChannelType:       c.ChannelType,
		Timestamp:         c.Timestamp,
		PreviousTimestamp: c.PreviousTimestamp,
		Date:              c.Date,
		EventParams:       toParams(c.EventParams),
		UserID:            c.UserID,
		UserPseudoID:      c.UserPseudoID,
		UserParams:        toParams(c.UserParams),
		Device:            toDevice(c.Device),
		AppInfo:           toAppInfo(c.AppInfo),
		Items:             toItems(c.Items),
	}
}

func toParams(dtos []ParamDTO) []domain.Param {
	params := make([]domain.Param, len(dtos))
	for i, dto := range dtos {
		params[i] = domain.Param{
			Key:          dto.Key,
			StringValue:  dto.StringValue,
			NumberValue:  dto.NumberValue,
			BooleanValue: dto.BooleanValue,
		}
	}
	return params
}

func toDevice(dto DeviceDTO) domain.Device {
	return domain.Device{
		Category:               dto.Category,
		MobileBrandName:        dto.MobileBrandName,
		MobileModelName:        dto.MobileModelName,
		OperatingSystem:        dto.OperatingSystem,
		OperatingSystemVersion: dto.OperatingSystemVersion,
		Language:               dto.Language,
		BrowserName:            dto.BrowserName,
		BrowserVersion:         dto.BrowserVersion,
		Hostname:               dto.Hostname,
	}
}

func toAppInfo(dto AppInfoDTO) domain.AppInfo {
	return domain.AppInfo{
		ID:      dto.ID,
		Version: dto.Version,
	}
}

func toItems(dtos []ItemDTO) []domain.Item {
	items := make([]domain.Item, len(dtos))
	for i, dto := range dtos {
		items[i] = domain.Item{
			ID:            dto.ID,
			Name:          dto.Name,
			Brand:         dto.Brand,
			Variant:       dto.Variant,
			PriceInUsd:    dto.PriceInUsd,
			Quantity:      dto.Quantity,
			RevenueInUsd:  dto.RevenueInUsd,
			LocationId:    dto.LocationId,
			ListId:        dto.ListId,
			ListName:      dto.ListName,
			PromotionId:   dto.PromotionId,
			PromotionName: dto.PromotionName,
			Params:        toParams(dto.Params),
		}
	}
	return items
}
