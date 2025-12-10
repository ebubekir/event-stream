package clickhouse

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Migration represents a single migration file
type Migration struct {
	Version  string
	Name     string
	UpSQL    string
	DownSQL  string
	Filename string
}

// Migrator handles database migrations for ClickHouse
type Migrator struct {
	db         *ClickHouseDb
	migrations []Migration
	tableName  string
}

// NewMigrator creates a new Migrator instance from embedded filesystem
func NewMigrator(db *ClickHouseDb, migrationFS embed.FS) (*Migrator, error) {
	m := &Migrator{
		db:        db,
		tableName: "schema_migrations",
	}

	if err := m.loadMigrations(migrationFS); err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	return m, nil
}

// loadMigrations reads all migration files from the embedded filesystem
func (m *Migrator) loadMigrations(migrationFS embed.FS) error {
	// Regex to parse migration filenames: 000001_create_events_table.up.sql
	re := regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

	migrationMap := make(map[string]*Migration)

	err := fs.WalkDir(migrationFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		filename := d.Name()
		matches := re.FindStringSubmatch(filename)
		if matches == nil {
			return nil // Skip non-migration files
		}

		version := matches[1]
		name := matches[2]
		direction := matches[3]

		content, err := fs.ReadFile(migrationFS, path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		key := version + "_" + name
		if _, exists := migrationMap[key]; !exists {
			migrationMap[key] = &Migration{
				Version:  version,
				Name:     name,
				Filename: key,
			}
		}

		if direction == "up" {
			migrationMap[key].UpSQL = string(content)
		} else {
			migrationMap[key].DownSQL = string(content)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Convert map to sorted slice
	for _, migration := range migrationMap {
		m.migrations = append(m.migrations, *migration)
	}

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	return nil
}

// ensureMigrationTable creates the schema_migrations table if it doesn't exist
func (m *Migrator) ensureMigrationTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version     String,
			name        String,
			applied_at  DateTime DEFAULT now()
		) ENGINE = MergeTree()
		ORDER BY version
	`, m.tableName)

	return ExecWithContext(ctx, m.db, query)
}

// getAppliedMigrations returns a set of applied migration versions
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	type migrationRecord struct {
		Version string `db:"version"`
	}

	var records []migrationRecord
	query := fmt.Sprintf("SELECT version FROM %s", m.tableName)

	if err := SelectWithContext(ctx, m.db, &records, query); err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	applied := make(map[string]bool)
	for _, r := range records {
		applied[r.Version] = true
	}

	return applied, nil
}

// recordMigration marks a migration as applied
func (m *Migrator) recordMigration(ctx context.Context, migration Migration) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (version, name) VALUES (?, ?)",
		m.tableName,
	)
	return ExecWithContext(ctx, m.db, query, migration.Version, migration.Name)
}

// removeMigrationRecord removes a migration record (for rollback)
func (m *Migrator) removeMigrationRecord(ctx context.Context, version string) error {
	query := fmt.Sprintf(
		"ALTER TABLE %s DELETE WHERE version = ?",
		m.tableName,
	)
	return ExecWithContext(ctx, m.db, query, version)
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureMigrationTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	for _, migration := range m.migrations {
		if applied[migration.Version] {
			log.Printf("[migrate] Skipping %s (already applied)", migration.Filename)
			continue
		}

		if migration.UpSQL == "" {
			log.Printf("[migrate] Skipping %s (no up migration)", migration.Filename)
			continue
		}

		log.Printf("[migrate] Applying %s...", migration.Filename)

		// Execute each statement separately (ClickHouse doesn't support multi-statement in one exec)
		statements := splitStatements(migration.UpSQL)
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if err := ExecWithContext(ctx, m.db, stmt); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w\nStatement: %s", migration.Filename, err, stmt)
			}
		}

		if err := m.recordMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Filename, err)
		}

		log.Printf("[migrate] Applied %s successfully", migration.Filename)
	}

	return nil
}

// Down rolls back the last applied migration
func (m *Migrator) Down(ctx context.Context) error {
	if err := m.ensureMigrationTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Find the last applied migration
	var lastApplied *Migration
	for i := len(m.migrations) - 1; i >= 0; i-- {
		if applied[m.migrations[i].Version] {
			lastApplied = &m.migrations[i]
			break
		}
	}

	if lastApplied == nil {
		log.Println("[migrate] No migrations to rollback")
		return nil
	}

	if lastApplied.DownSQL == "" {
		return fmt.Errorf("migration %s has no down migration", lastApplied.Filename)
	}

	log.Printf("[migrate] Rolling back %s...", lastApplied.Filename)

	statements := splitStatements(lastApplied.DownSQL)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if err := ExecWithContext(ctx, m.db, stmt); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w\nStatement: %s", lastApplied.Filename, err, stmt)
		}
	}

	if err := m.removeMigrationRecord(ctx, lastApplied.Version); err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", lastApplied.Filename, err)
	}

	// ClickHouse DELETE is async, wait a bit for consistency
	time.Sleep(100 * time.Millisecond)

	log.Printf("[migrate] Rolled back %s successfully", lastApplied.Filename)
	return nil
}

// DownAll rolls back all applied migrations
func (m *Migrator) DownAll(ctx context.Context) error {
	if err := m.ensureMigrationTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Rollback in reverse order
	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]

		if !applied[migration.Version] {
			continue
		}

		if migration.DownSQL == "" {
			return fmt.Errorf("migration %s has no down migration", migration.Filename)
		}

		log.Printf("[migrate] Rolling back %s...", migration.Filename)

		statements := splitStatements(migration.DownSQL)
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if err := ExecWithContext(ctx, m.db, stmt); err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w\nStatement: %s", migration.Filename, err, stmt)
			}
		}

		if err := m.removeMigrationRecord(ctx, migration.Version); err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", migration.Filename, err)
		}

		log.Printf("[migrate] Rolled back %s successfully", migration.Filename)
	}

	// ClickHouse DELETE is async, wait a bit for consistency
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Status returns the current migration status
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := m.ensureMigrationTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	var statuses []MigrationStatus
	for _, migration := range m.migrations {
		statuses = append(statuses, MigrationStatus{
			Version: migration.Version,
			Name:    migration.Name,
			Applied: applied[migration.Version],
		})
	}

	return statuses, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version string
	Name    string
	Applied bool
}

// splitStatements splits SQL content into individual statements
func splitStatements(sql string) []string {
	// Remove comments and split by semicolon
	var statements []string
	var current strings.Builder
	inComment := false

	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		// Handle multi-line comments
		if strings.Contains(trimmed, "/*") {
			inComment = true
		}
		if strings.Contains(trimmed, "*/") {
			inComment = false
			continue
		}
		if inComment {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		// Check if statement ends with semicolon
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			stmt = strings.TrimSuffix(stmt, ";")
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	// Handle last statement without semicolon
	if stmt := strings.TrimSpace(current.String()); stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}
