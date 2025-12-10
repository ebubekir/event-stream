package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jmoiron/sqlx"
)

// ClickHouseDb wraps sqlx.DB for ClickHouse connection management
type ClickHouseDb struct {
	ConnectionString string
	Database         string // ClickHouse database name
	DB               *sqlx.DB
}

// New creates a new ClickHouseDb instance with the given connection string and database
func New(connectionString, database string) *ClickHouseDb {
	return &ClickHouseDb{
		ConnectionString: connectionString,
		Database:         database,
		DB:               nil,
	}
}

// NewWithOptions creates a new ClickHouseDb instance with ClickHouse-specific options
func NewWithOptions(host string, port int, database, username, password string) *ClickHouseDb {
	connStr := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s",
		username, password, host, port, database)
	return &ClickHouseDb{
		ConnectionString: connStr,
		Database:         database,
		DB:               nil,
	}
}

// getDB returns the sqlx.DB instance, initializing it if necessary
func (c *ClickHouseDb) getDB() (*sqlx.DB, error) {
	if c.DB != nil {
		return c.DB, nil
	}

	opts, err := clickhouse.ParseDSN(c.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ClickHouse DSN: %w", err)
	}

	// Set default database if provided
	if c.Database != "" {
		opts.Auth.Database = c.Database
	}

	// Create native connection
	conn := clickhouse.OpenDB(opts)

	// Wrap with sqlx for convenience
	db := sqlx.NewDb(conn, "clickhouse")

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	c.DB = db
	return db, nil
}

// CheckConnection verifies the database connection
func (c *ClickHouseDb) CheckConnection() error {
	db, err := c.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

// Close closes the database connection
func (c *ClickHouseDb) Close() error {
	if c.DB == nil {
		return nil
	}
	err := c.DB.Close()
	c.DB = nil
	return err
}

// Get retrieves a single row and scans it into dest (type-safe with sqlx)
func Get[T any](db *ClickHouseDb, dest *T, query string, args ...interface{}) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return sqlxDB.GetContext(ctx, dest, query, args...)
}

// Select retrieves multiple rows and scans them into dest (type-safe with sqlx)
func Select[T any](db *ClickHouseDb, dest *[]T, query string, args ...interface{}) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return sqlxDB.SelectContext(ctx, dest, query, args...)
}

// Exec executes a query without returning rows (INSERT, ALTER, etc.)
func Exec(db *ClickHouseDb, query string, args ...interface{}) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = sqlxDB.ExecContext(ctx, query, args...)
	return err
}

// NamedExec executes a named query (supports :param syntax)
func NamedExec(db *ClickHouseDb, query string, arg interface{}) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = sqlxDB.NamedExecContext(ctx, query, arg)
	return err
}

// NamedGet retrieves a single row using named parameters
func NamedGet[T any](db *ClickHouseDb, dest *T, query string, arg interface{}) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := sqlxDB.NamedQueryContext(ctx, query, arg)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return sql.ErrNoRows
	}

	return rows.StructScan(dest)
}

// Count returns the number of rows matching the query
func Count(db *ClickHouseDb, query string, args ...interface{}) (uint64, error) {
	sqlxDB, err := db.getDB()
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var count uint64
	err = sqlxDB.GetContext(ctx, &count, query, args...)
	return count, err
}

// BatchInsert performs a batch insert for better performance
// ClickHouse is optimized for batch operations
func BatchInsert[T any](db *ClickHouseDb, query string, items []T) error {
	if len(items) == 0 {
		return nil
	}

	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	tx, err := sqlxDB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin batch transaction: %w", err)
	}

	for _, item := range items {
		_, err := tx.NamedExecContext(ctx, query, item)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("batch insert failed: %w", err)
		}
	}

	return tx.Commit()
}

// ExecWithContext executes a query with custom context
func ExecWithContext(ctx context.Context, db *ClickHouseDb, query string, args ...interface{}) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	_, err = sqlxDB.ExecContext(ctx, query, args...)
	return err
}

// SelectWithContext retrieves multiple rows with custom context
func SelectWithContext[T any](ctx context.Context, db *ClickHouseDb, dest *[]T, query string, args ...interface{}) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	return sqlxDB.SelectContext(ctx, dest, query, args...)
}

// GetDatabase returns the database name
func (c *ClickHouseDb) GetDatabase() string {
	return c.Database
}

// GetRawDB returns the underlying sqlx.DB instance
// Use this when you need direct access to sqlx functionality
func (c *ClickHouseDb) GetRawDB() (*sqlx.DB, error) {
	return c.getDB()
}
