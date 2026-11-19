package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	maxOpenConnections = 25
	maxIdleConnections = 5
	connectionLifetime = 30 * time.Minute
	connectionIdleTime = 5 * time.Minute
	connectionTimeout  = 5 * time.Second
)

func Open(ctx context.Context, databaseURL string) (*sql.DB, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure the connection pool.
	db.SetMaxOpenConns(maxOpenConnections)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxLifetime(connectionLifetime)
	db.SetConnMaxIdleTime(connectionIdleTime)

	// Verify that PostgreSQL is reachable.
	pingCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
