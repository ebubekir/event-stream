package event

import (
	"context"
	"time"
)

// AggregationType defines how metrics should be grouped
type AggregationType string

const (
	AggregationByChannel AggregationType = "channel"
	AggregationByDaily   AggregationType = "daily"
	AggregationByHourly  AggregationType = "hourly"
)

// MetricsQuery represents the query parameters for fetching metrics
type MetricsQuery struct {
	EventName   string
	From        time.Time
	To          time.Time
	Aggregation AggregationType
}

// GroupedMetric represents metrics for a specific group
type GroupedMetric struct {
	GroupKey        string
	TotalCount      int64
	UniqueUserCount int64
}

// MetricsResult represents the result of a metrics query
type MetricsResult struct {
	EventName       string
	From            time.Time
	To              time.Time
	TotalCount      int64
	UniqueUserCount int64
	GroupedMetrics  []GroupedMetric
}

// EventMetricsReader defines the contract for reading event metrics
// This interface lives in domain layer - implementations in adapter/outbound
type EventMetricsReader interface {
	// GetMetrics retrieves aggregated metrics for events matching the query
	GetMetrics(ctx context.Context, query *MetricsQuery) (*MetricsResult, error)
}
