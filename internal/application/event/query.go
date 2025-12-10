package event

import (
	"time"

	eventDomain "github.com/ebubekir/event-stream/internal/domain/event"
)

// GetMetricsQuery represents the query for fetching event metrics
type GetMetricsQuery struct {
	EventName   string
	From        time.Time
	To          time.Time
	Aggregation string // "channel", "daily", "hourly"
}

// ToMetricsQuery converts application query to domain query
func (q *GetMetricsQuery) ToMetricsQuery() *eventDomain.MetricsQuery {
	return &eventDomain.MetricsQuery{
		EventName:   q.EventName,
		From:        q.From,
		To:          q.To,
		Aggregation: eventDomain.AggregationType(q.Aggregation),
	}
}

// GroupedMetricDTO represents metrics for a specific group in application layer
type GroupedMetricDTO struct {
	GroupKey        string
	TotalCount      int64
	UniqueUserCount int64
}

// MetricsResultDTO represents the metrics result in application layer
type MetricsResultDTO struct {
	EventName       string
	From            time.Time
	To              time.Time
	TotalCount      int64
	UniqueUserCount int64
	GroupedMetrics  []GroupedMetricDTO
}

// FromMetricsResult converts domain result to application DTO
func FromMetricsResult(result *eventDomain.MetricsResult) *MetricsResultDTO {
	groupedMetrics := make([]GroupedMetricDTO, len(result.GroupedMetrics))
	for i, gm := range result.GroupedMetrics {
		groupedMetrics[i] = GroupedMetricDTO{
			GroupKey:        gm.GroupKey,
			TotalCount:      gm.TotalCount,
			UniqueUserCount: gm.UniqueUserCount,
		}
	}

	return &MetricsResultDTO{
		EventName:       result.EventName,
		From:            result.From,
		To:              result.To,
		TotalCount:      result.TotalCount,
		UniqueUserCount: result.UniqueUserCount,
		GroupedMetrics:  groupedMetrics,
	}
}
