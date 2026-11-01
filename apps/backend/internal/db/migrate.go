package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

// Migrations are append-only and generated from schema/veil.schema:
// `openschema migrate schema/veil.schema -t migration-sqlite --numbered
// apps/backend/internal/db/migrations --name <name>` writes the next file.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

// migrate applies every embedded migration not yet recorded in _migration,
// each in its own transaction. Generated migrations rebuild tables, which
// needs foreign keys off, and SQLite ignores that pragma inside a
// transaction, so the whole run happens on one connection with them off.
func (d *DB) migrate(ctx context.Context) error {
	if _, err := d.pool.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS _migration (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at TEXT NOT NULL
	)`); err != nil {
		return fmt.Errorf("create _migration: %w", err)
	}
	pending, err := d.pendingMigrations(ctx)
	if err != nil || len(pending) == 0 {
		return err
	}
	connection, err := d.pool.Conn(ctx)
	if err != nil {
		return err
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return err
	}
	defer connection.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	for _, name := range pending {
		if err := applyMigration(ctx, connection, name); err != nil {
			return err
		}
		log.Printf("db: applied migration %s", name)
	}
	return checkForeignKeys(ctx, connection)
}

// pendingMigrations lists the embedded migrations not yet applied, in order.
func (d *DB) pendingMigrations(ctx context.Context) ([]string, error) {
	applied, err := d.appliedMigrations(ctx)
	if err != nil {
		return nil, err
	}
	names, err := migrationNames()
	if err != nil {
		return nil, err
	}
	pending := []string{}
	for _, name := range names {
		version, err := migrationVersion(name)
		if err != nil {
			return nil, err
		}
		if !applied[version] {
			pending = append(pending, name)
		}
	}
	return pending, nil
}

// checkForeignKeys fails when the migrations left rows pointing at rows that
// don't exist.
func checkForeignKeys(ctx context.Context, connection *sql.Conn) error {
	rows, err := connection.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return fmt.Errorf("migrations left broken foreign keys")
	}
	return rows.Err()
}

// appliedMigrations returns the set of recorded migration versions.
func (d *DB) appliedMigrations(ctx context.Context) (map[int]bool, error) {
	rows, err := d.pool.QueryContext(ctx, "SELECT version FROM _migration")
	if err != nil {
		return nil, fmt.Errorf("read _migration: %w", err)
	}
	defer rows.Close()
	applied := map[int]bool{}
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

// applyMigration runs one migration file and records it, atomically.
func applyMigration(ctx context.Context, connection *sql.Conn, name string) error {
	version, err := migrationVersion(name)
	if err != nil {
		return err
	}
	script, err := migrationFiles.ReadFile("migrations/" + name)
	if err != nil {
		return err
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := transaction.ExecContext(ctx, string(script)); err != nil {
		transaction.Rollback()
		return fmt.Errorf("migration %s: %w", name, err)
	}
	if _, err := transaction.ExecContext(ctx,
		"INSERT INTO _migration (version, name, applied_at) VALUES (?, ?, ?)", version, name, Now()); err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

// migrationNames lists the embedded migration files in version order.
func migrationNames() ([]string, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

// migrationVersion parses the NNN prefix of "NNN_name.sql".
func migrationVersion(name string) (int, error) {
	prefix, _, found := strings.Cut(name, "_")
	if !found {
		return 0, fmt.Errorf("migration %q: expected NNN_name.sql", name)
	}
	version, err := strconv.Atoi(prefix)
	if err != nil {
		return 0, fmt.Errorf("migration %q: %w", name, err)
	}
	return version, nil
}
