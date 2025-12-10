package clickhouse

import (
	"context"
	"fmt"

	eventDomain "github.com/ebubekir/event-stream/internal/domain/event"
	"github.com/ebubekir/event-stream/pkg/clickhouse"
)

// MetricsReader implements domain/event.EventMetricsReader for ClickHouse
type MetricsReader struct {
	db *clickhouse.ClickHouseDb
}

// NewMetricsReader creates a new ClickHouse metrics reader
func NewMetricsReader(db *clickhouse.ClickHouseDb) *MetricsReader {
	return &MetricsReader{db: db}
}

// metricsRow represents a row from the metrics query
type metricsRow struct {
	TotalCount      int64 `db:"total_count"`
	UniqueUserCount int64 `db:"unique_user_count"`
}

// groupedMetricsRow represents a row from the grouped metrics query
type groupedMetricsRow struct {
	GroupKey        string `db:"group_key"`
	TotalCount      int64  `db:"total_count"`
	UniqueUserCount int64  `db:"unique_user_count"`
}

// GetMetrics retrieves aggregated metrics for events matching the query
func (r *MetricsReader) GetMetrics(ctx context.Context, query *eventDomain.MetricsQuery) (*eventDomain.MetricsResult, error) {
	result := &eventDomain.MetricsResult{
		EventName: query.EventName,
		From:      query.From,
		To:        query.To,
	}

	// Build base WHERE clause
	whereClause := "WHERE name = ?"
	args := []interface{}{query.EventName}

	if !query.From.IsZero() {
		whereClause += " AND date >= ?"
		args = append(args, query.From)
	}

	if !query.To.IsZero() {
		whereClause += " AND date <= ?"
		args = append(args, query.To)
	}

	// Get totals
	totalsQuery := fmt.Sprintf(`
		SELECT 
			count() AS total_count,
			uniqExact(user_id) AS unique_user_count
		FROM events
		%s
	`, whereClause)

	var totals metricsRow
	if err := clickhouse.Get(r.db, &totals, totalsQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to query totals: %w", err)
	}

	result.TotalCount = totals.TotalCount
	result.UniqueUserCount = totals.UniqueUserCount

	// Get grouped metrics if aggregation is specified
	if query.Aggregation != "" {
		groupedMetrics, err := r.getGroupedMetrics(ctx, query, whereClause, args)
		if err != nil {
			return nil, err
		}
		result.GroupedMetrics = groupedMetrics
	}

	return result, nil
}

// getGroupedMetrics retrieves metrics grouped by the specified aggregation
func (r *MetricsReader) getGroupedMetrics(ctx context.Context, query *eventDomain.MetricsQuery, whereClause string, args []interface{}) ([]eventDomain.GroupedMetric, error) {
	var groupByExpr string
	var selectExpr string

	switch query.Aggregation {
	case eventDomain.AggregationByChannel:
		groupByExpr = "channel_type"
		selectExpr = "channel_type AS group_key"
	case eventDomain.AggregationByDaily:
		groupByExpr = "toDate(date)"
		selectExpr = "toString(toDate(date)) AS group_key"
	case eventDomain.AggregationByHourly:
		groupByExpr = "toStartOfHour(date)"
		selectExpr = "toString(toStartOfHour(date)) AS group_key"
	default:
		return nil, nil
	}

	groupedQuery := fmt.Sprintf(`
		SELECT 
			%s,
			count() AS total_count,
			uniqExact(user_id) AS unique_user_count
		FROM events
		%s
		GROUP BY %s
		ORDER BY %s
	`, selectExpr, whereClause, groupByExpr, groupByExpr)

	var rows []groupedMetricsRow
	if err := clickhouse.SelectWithContext(ctx, r.db, &rows, groupedQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to query grouped metrics: %w", err)
	}

	groupedMetrics := make([]eventDomain.GroupedMetric, len(rows))
	for i, row := range rows {
		groupedMetrics[i] = eventDomain.GroupedMetric{
			GroupKey:        row.GroupKey,
			TotalCount:      row.TotalCount,
			UniqueUserCount: row.UniqueUserCount,
		}
	}

	return groupedMetrics, nil
}
