package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/degeta10/workout-assistant-api/internal/config"
	_ "github.com/lib/pq"
)

const (
	pingAttempts       = 5
	pingTimeoutPerTry  = 5 * time.Second
	pingRetryBaseDelay = 300 * time.Millisecond
)

func InitDB(cfg config.DBConfig) (*sql.DB, error) {
	return InitDBWithContext(context.Background(), cfg)
}

func InitDBWithContext(parentCtx context.Context, cfg config.DBConfig) (*sql.DB, error) {
	if parentCtx == nil {
		parentCtx = context.Background()
	}

	if err := validateDBConfig(cfg); err != nil {
		return nil, err
	}

	// 1. Prepare credentials
	encodedPass := url.QueryEscape(cfg.Password)

	// 2. Build the DSN (Standard format)
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require",
		cfg.User, encodedPass, cfg.Host, cfg.Port, cfg.Name)

	// 3. Open connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// 4. Heavy Duty Lambda Settings
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 5. Startup ping with retries for transient cold-start/network hiccups
	var pingErr error
	for attempt := 1; attempt <= pingAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(parentCtx, pingTimeoutPerTry)
		pingErr = db.PingContext(ctx)
		cancel()

		if pingErr == nil {
			return db, nil
		}

		if parentCtx.Err() != nil {
			_ = db.Close()
			return nil, fmt.Errorf("db init canceled: %w", parentCtx.Err())
		}

		if attempt < pingAttempts {
			time.Sleep(time.Duration(attempt) * pingRetryBaseDelay)
		}
	}

	_ = db.Close()
	return nil, fmt.Errorf("ping db after %d attempts: %w", pingAttempts, pingErr)
}

func validateDBConfig(cfg config.DBConfig) error {
	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("db config invalid: DB_HOST is empty")
	}
	if strings.TrimSpace(cfg.Port) == "" {
		return fmt.Errorf("db config invalid: DB_PORT is empty")
	}
	if strings.TrimSpace(cfg.User) == "" {
		return fmt.Errorf("db config invalid: DB_USER is empty")
	}
	if strings.TrimSpace(cfg.Password) == "" {
		return fmt.Errorf("db config invalid: DB_PASSWORD is empty")
	}
	if strings.TrimSpace(cfg.Name) == "" {
		return fmt.Errorf("db config invalid: DB_NAME is empty")
	}

	return nil
}
