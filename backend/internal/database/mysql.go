// Package database owns the MySQL connection lifecycle. It opens a pooled
// *sql.DB from configuration and verifies connectivity. No business queries
// live here — those belong to the repository layer.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver, registered via blank import.

	"iclassroom/backend/internal/config"
)

// New opens a MySQL connection pool, applies the configured pool limits, and
// pings the server to confirm the connection works. The caller owns the
// returned *sql.DB and must Close it on shutdown.
func New(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DBDSN())
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return db, nil
}
