package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/degeta10/workout-assistant-api/internal/config"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(cfg config.DBConfig) error {

	// 1. URL Encode Password (handles special chars)
	encodedPass := url.QueryEscape(cfg.Password)

	// 2. Build Connection String
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		cfg.User, encodedPass, cfg.Host, cfg.Port, cfg.Name)

	// 3. Log Connection Attempt (without password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// 4. INDUSTRY STANDARD: Pool Management
	// These settings prevent your Lambda from killing the DB
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 5. Verify
	if err := db.Ping(); err != nil {
		return fmt.Errorf("database unreachable: %v", err)
	}

	DB = db
	log.Println("Database Connection Pool Initialized")
	return nil
}
