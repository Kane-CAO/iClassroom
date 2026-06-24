package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ApplyMigrations applies local *.up.sql files once. It intentionally stays
// small because this project only needs deterministic local/demo migrations.
func ApplyMigrations(db *sql.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("database: list migrations: %w", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    VARCHAR(255) NOT NULL,
    applied_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`); err != nil {
		return fmt.Errorf("database: ensure schema_migrations: %w", err)
	}

	for _, file := range files {
		version := migrationVersion(file)
		applied, err := migrationApplied(ctx, db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		markOnly, err := migrationAlreadyInSchema(ctx, db, version)
		if err != nil {
			return err
		}
		if markOnly {
			if err := recordMigration(ctx, db, version); err != nil {
				return err
			}
			continue
		}

		body, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("database: read migration %s: %w", filepath.Base(file), err)
		}
		if err := execMigration(ctx, db, string(body)); err != nil {
			return fmt.Errorf("database: apply migration %s: %w", filepath.Base(file), err)
		}
		if err := recordMigration(ctx, db, version); err != nil {
			return err
		}
	}
	return nil
}

func migrationVersion(file string) string {
	return strings.TrimSuffix(filepath.Base(file), ".up.sql")
}

func migrationApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&count); err != nil {
		return false, fmt.Errorf("database: check migration %s: %w", version, err)
	}
	return count > 0, nil
}

func recordMigration(ctx context.Context, db *sql.DB, version string) error {
	if _, err := db.ExecContext(ctx, `INSERT IGNORE INTO schema_migrations (version) VALUES (?)`, version); err != nil {
		return fmt.Errorf("database: record migration %s: %w", version, err)
	}
	return nil
}

func migrationAlreadyInSchema(ctx context.Context, db *sql.DB, version string) (bool, error) {
	switch version {
	case "000001_init_schema":
		return tableExists(ctx, db, "rooms")
	case "000002_v2_accounts_attachments":
		admins, err := tableExists(ctx, db, "admin_users")
		if err != nil {
			return false, err
		}
		teachers, err := tableExists(ctx, db, "teacher_accounts")
		if err != nil {
			return false, err
		}
		sessions, err := tableExists(ctx, db, "auth_sessions")
		if err != nil {
			return false, err
		}
		attachments, err := tableExists(ctx, db, "submission_attachments")
		if err != nil {
			return false, err
		}
		teacherID, err := columnExists(ctx, db, "rooms", "teacher_id")
		if err != nil {
			return false, err
		}
		return admins && teachers && sessions && attachments && teacherID, nil
	default:
		return false, nil
	}
}

func tableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema = DATABASE() AND table_name = ?`, table).Scan(&count); err != nil {
		return false, fmt.Errorf("database: check table %s: %w", table, err)
	}
	return count > 0, nil
}

func columnExists(ctx context.Context, db *sql.DB, table, column string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM information_schema.columns
WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`, table, column).Scan(&count); err != nil {
		return false, fmt.Errorf("database: check column %s.%s: %w", table, column, err)
	}
	return count > 0, nil
}

func execMigration(ctx context.Context, db *sql.DB, body string) error {
	for _, stmt := range splitSQLStatements(body) {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func splitSQLStatements(body string) []string {
	parts := strings.Split(body, ";")
	stmts := make([]string, 0, len(parts))
	for _, part := range parts {
		if stmt := strings.TrimSpace(part); stmt != "" {
			stmts = append(stmts, stmt)
		}
	}
	return stmts
}
