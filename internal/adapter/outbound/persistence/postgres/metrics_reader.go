package postgres

import (
	"context"
	"fmt"

	eventDomain "github.com/ebubekir/event-stream/internal/domain/event"
	"github.com/ebubekir/event-stream/pkg/postgresql"
)

// MetricsReader implements domain/event.EventMetricsReader for PostgreSQL
type MetricsReader struct {
	db *postgresql.PostgresDb
}

// NewMetricsReader creates a new PostgreSQL metrics reader
func NewMetricsReader(db *postgresql.PostgresDb) *MetricsReader {
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
	whereClause := "WHERE name = $1"
	args := []interface{}{query.EventName}
	argIndex := 2

	if !query.From.IsZero() {
		whereClause += fmt.Sprintf(" AND date >= $%d", argIndex)
		args = append(args, query.From)
		argIndex++
	}

	if !query.To.IsZero() {
		whereClause += fmt.Sprintf(" AND date <= $%d", argIndex)
		args = append(args, query.To)
	}

	// Get totals
	totalsQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) AS total_count,
			COUNT(DISTINCT user_id) AS unique_user_count
		FROM events
		%s
	`, whereClause)

	var totals metricsRow
	if err := postgresql.Get(r.db, &totals, totalsQuery, args...); err != nil {
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
		groupByExpr = "DATE(date)"
		selectExpr = "TO_CHAR(DATE(date), 'YYYY-MM-DD') AS group_key"
	case eventDomain.AggregationByHourly:
		groupByExpr = "DATE_TRUNC('hour', date)"
		selectExpr = "TO_CHAR(DATE_TRUNC('hour', date), 'YYYY-MM-DD HH24:00:00') AS group_key"
	default:
		return nil, nil
	}

	groupedQuery := fmt.Sprintf(`
		SELECT 
			%s,
			COUNT(*) AS total_count,
			COUNT(DISTINCT user_id) AS unique_user_count
		FROM events
		%s
		GROUP BY %s
		ORDER BY %s
	`, selectExpr, whereClause, groupByExpr, groupByExpr)

	var rows []groupedMetricsRow
	if err := postgresql.Select(r.db, &rows, groupedQuery, args...); err != nil {
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
