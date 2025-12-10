package dto

import (
	"github.com/ebubekir/event-stream/internal/application/event"
	"github.com/ebubekir/event-stream/internal/domain"
)

// CreateEventRequest represents the HTTP request body for creating an event
type CreateEventRequest struct {
	Name              string         `json:"name" binding:"required"`
	ChannelType       string         `json:"channel_type" binding:"required,oneof=web mobile desktop tv console other"`
	Timestamp         int64          `json:"timestamp"`
	PreviousTimestamp int64          `json:"previous_timestamp"`
	Date              string         `json:"date"`
	EventParams       []ParamRequest `json:"event_params"`
	UserID            string         `json:"user_id"`
	UserPseudoID      string         `json:"user_pseudo_id"`
	UserParams        []ParamRequest `json:"user_params"`
	Device            DeviceRequest  `json:"device"`
	AppInfo           AppInfoRequest `json:"app_info"`
	Items             []ItemRequest  `json:"items"`
} // @name CreateEventRequest

// ParamRequest represents a parameter in HTTP request
type ParamRequest struct {
	Key          string  `json:"key" binding:"required"`
	StringValue  string  `json:"string_value"`
	NumberValue  float64 `json:"number_value"`
	BooleanValue bool    `json:"boolean_value"`
} // @name ParamRequest

// DeviceRequest represents device information in HTTP request
type DeviceRequest struct {
	Category               string `json:"category"`
	MobileBrandName        string `json:"mobile_brand_name"`
	MobileModelName        string `json:"mobile_model_name"`
	OperatingSystem        string `json:"operating_system"`
	OperatingSystemVersion string `json:"operating_system_version"`
	Language               string `json:"language"`
	BrowserName            string `json:"browser_name"`
	BrowserVersion         string `json:"browser_version"`
	Hostname               string `json:"hostname"`
} // @name DeviceRequest

// AppInfoRequest represents app information in HTTP request
type AppInfoRequest struct {
	ID      string `json:"id"`
	Version string `json:"version"`
} // @name AppInfoRequest

// ItemRequest represents an item in HTTP request
type ItemRequest struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Brand         string         `json:"brand"`
	Variant       string         `json:"variant"`
	PriceInUsd    float64        `json:"price_in_usd"`
	Quantity      int            `json:"quantity"`
	RevenueInUsd  float64        `json:"revenue_in_usd"`
	LocationId    string         `json:"location_id"`
	ListId        string         `json:"list_id"`
	ListName      string         `json:"list_name"`
	PromotionId   string         `json:"promotion_id"`
	PromotionName string         `json:"promotion_name"`
	Params        []ParamRequest `json:"params"`
} // @name ItemRequest

// CreateEventResponse represents the HTTP response after creating an event
type CreateEventResponse struct {
	ID string `json:"id"`
} // @name CreateEventResponse

// CreateEventBatchRequest represents batch event creation request
type CreateEventBatchRequest struct {
	Events []CreateEventRequest `json:"events" binding:"required,dive"`
} // @name CreateEventBatchRequest

// CreateEventBatchResponse represents batch event creation response
type CreateEventBatchResponse struct {
	IDs []string `json:"ids"`
} // @name CreateEventBatchResponse

// ToCommand converts HTTP DTO to application command
func (r *CreateEventRequest) ToCommand() *event.CreateEventCommand {
	return &event.CreateEventCommand{
		Name:              r.Name,
		ChannelType:       domain.ChannelType(r.ChannelType),
		Timestamp:         r.Timestamp,
		PreviousTimestamp: r.PreviousTimestamp,
		Date:              r.Date,
		EventParams:       toParamDTOs(r.EventParams),
		UserID:            r.UserID,
		UserPseudoID:      r.UserPseudoID,
		UserParams:        toParamDTOs(r.UserParams),
		Device:            toDeviceDTO(r.Device),
		AppInfo:           toAppInfoDTO(r.AppInfo),
		Items:             toItemDTOs(r.Items),
	}
}

func toParamDTOs(requests []ParamRequest) []event.ParamDTO {
	params := make([]event.ParamDTO, len(requests))
	for i, req := range requests {
		params[i] = event.ParamDTO{
			Key:          req.Key,
			StringValue:  req.StringValue,
			NumberValue:  req.NumberValue,
			BooleanValue: req.BooleanValue,
		}
	}
	return params
}

func toDeviceDTO(req DeviceRequest) event.DeviceDTO {
	return event.DeviceDTO{
		Category:               req.Category,
		MobileBrandName:        req.MobileBrandName,
		MobileModelName:        req.MobileModelName,
		OperatingSystem:        req.OperatingSystem,
		OperatingSystemVersion: req.OperatingSystemVersion,
		Language:               req.Language,
		BrowserName:            req.BrowserName,
		BrowserVersion:         req.BrowserVersion,
		Hostname:               req.Hostname,
	}
}

func toAppInfoDTO(req AppInfoRequest) event.AppInfoDTO {
	return event.AppInfoDTO{
		ID:      req.ID,
		Version: req.Version,
	}
}

func toItemDTOs(requests []ItemRequest) []event.ItemDTO {
	items := make([]event.ItemDTO, len(requests))
	for i, req := range requests {
		items[i] = event.ItemDTO{
			ID:            req.ID,
			Name:          req.Name,
			Brand:         req.Brand,
			Variant:       req.Variant,
			PriceInUsd:    req.PriceInUsd,
			Quantity:      req.Quantity,
			RevenueInUsd:  req.RevenueInUsd,
			LocationId:    req.LocationId,
			ListId:        req.ListId,
			ListName:      req.ListName,
			PromotionId:   req.PromotionId,
			PromotionName: req.PromotionName,
			Params:        toParamDTOs(req.Params),
		}
	}
	return items
}
