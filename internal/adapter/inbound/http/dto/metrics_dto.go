package dto

import (
	"time"

	"github.com/ebubekir/event-stream/internal/application/event"
)

// GetMetricsRequest represents the HTTP query parameters for metrics
type GetMetricsRequest struct {
	EventName   string `form:"event_name" binding:"required"`
	From        string `form:"from"`                                                    // RFC3339 format
	To          string `form:"to"`                                                      // RFC3339 format
	Aggregation string `form:"group_by" binding:"omitempty,oneof=channel daily hourly"` // channel, daily, hourly
} // @name GetMetricsRequest

// ToQuery converts HTTP request to application query
func (r *GetMetricsRequest) ToQuery() (*event.GetMetricsQuery, error) {
	query := &event.GetMetricsQuery{
		EventName:   r.EventName,
		Aggregation: r.Aggregation,
	}

	// Parse 'from' timestamp
	if r.From != "" {
		from, err := time.Parse(time.RFC3339, r.From)
		if err != nil {
			return nil, err
		}
		query.From = from
	}

	// Parse 'to' timestamp
	if r.To != "" {
		to, err := time.Parse(time.RFC3339, r.To)
		if err != nil {
			return nil, err
		}
		query.To = to
	} else {
		// Default to now if not provided
		query.To = time.Now()
	}

	return query, nil
}

// GroupedMetricResponse represents a grouped metric in the response
type GroupedMetricResponse struct {
	GroupKey        string `json:"group_key"`
	TotalCount      int64  `json:"total_count"`
	UniqueUserCount int64  `json:"unique_user_count"`
} // @name GroupedMetricResponse

// GetMetricsResponse represents the HTTP response for metrics
type GetMetricsResponse struct {
	EventName       string                  `json:"event_name"`
	From            string                  `json:"from"`
	To              string                  `json:"to"`
	TotalCount      int64                   `json:"total_count"`
	UniqueUserCount int64                   `json:"unique_user_count"`
	GroupedMetrics  []GroupedMetricResponse `json:"grouped_metrics,omitempty"`
} // @name GetMetricsResponse

// FromMetricsResultDTO converts application DTO to HTTP response
func FromMetricsResultDTO(dto *event.MetricsResultDTO) *GetMetricsResponse {
	groupedMetrics := make([]GroupedMetricResponse, len(dto.GroupedMetrics))
	for i, gm := range dto.GroupedMetrics {
		groupedMetrics[i] = GroupedMetricResponse{
			GroupKey:        gm.GroupKey,
			TotalCount:      gm.TotalCount,
			UniqueUserCount: gm.UniqueUserCount,
		}
	}

	return &GetMetricsResponse{
		EventName:       dto.EventName,
		From:            dto.From.Format(time.RFC3339),
		To:              dto.To.Format(time.RFC3339),
		TotalCount:      dto.TotalCount,
		UniqueUserCount: dto.UniqueUserCount,
		GroupedMetrics:  groupedMetrics,
	}
}
