package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresDb wraps sqlx.DB for connection management
type PostgresDb struct {
	ConnectionString string
	Schema           string // PostgreSQL schema name (e.g., "auth", "product")
	DB               *sqlx.DB
}

// New creates a new PostgresDb instance with the given connection string and schema
func New(connectionString, schema string) *PostgresDb {
	return &PostgresDb{
		ConnectionString: connectionString,
		Schema:           schema,
		DB:               nil,
	}
}

// getDB returns the sqlx.DB instance, initializing it if necessary
func (p *PostgresDb) getDB() (*sqlx.DB, error) {
	if p.DB != nil {
		return p.DB, nil
	}

	db, err := sqlx.Connect("postgres", p.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Set search_path to use the specified schema
	if p.Schema != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", p.Schema))
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set schema: %w", err)
		}
	}

	p.DB = db
	return db, nil
}

// CheckConnection verifies the database connection
func (p *PostgresDb) CheckConnection() error {
	db, err := p.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

// Close closes the database connection
func (p *PostgresDb) Close() error {
	if p.DB == nil {
		return nil
	}
	err := p.DB.Close()
	p.DB = nil
	return err
}

// Get retrieves a single row and scans it into dest (type-safe with sqlx)
func Get[T any](db *PostgresDb, dest *T, query string, args ...interface{}) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return sqlxDB.GetContext(ctx, dest, query, args...)
}

// Select retrieves multiple rows and scans them into dest (type-safe with sqlx)
func Select[T any](db *PostgresDb, dest *[]T, query string, args ...interface{}) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return sqlxDB.SelectContext(ctx, dest, query, args...)
}

// Exec executes a query without returning rows (INSERT, UPDATE, DELETE)
func Exec(db *PostgresDb, query string, args ...interface{}) error {
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
func NamedExec(db *PostgresDb, query string, arg interface{}) error {
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
func NamedGet[T any](db *PostgresDb, dest *T, query string, arg interface{}) error {
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
		return fmt.Errorf("no rows found")
	}

	return rows.StructScan(dest)
}

// Count returns the number of rows matching the query
func Count(db *PostgresDb, query string, args ...interface{}) (int64, error) {
	sqlxDB, err := db.getDB()
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var count int64
	err = sqlxDB.GetContext(ctx, &count, query, args...)
	return count, err
}

// Transaction executes a function within a database transaction
func Transaction(db *PostgresDb, fn func(*sqlx.Tx) error) error {
	sqlxDB, err := db.getDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := sqlxDB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	// Ensure schema is set for this transaction
	if db.Schema != "" {
		_, err = tx.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", db.Schema))
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to set schema in transaction: %w", err)
		}
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetSchema returns the schema name
func (p *PostgresDb) GetSchema() string {
	return p.Schema
}
