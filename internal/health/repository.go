package health

import (
	"context"
	"database/sql"
	"fmt"
)

type DatabaseRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &DatabaseRepository{db: db}
}

func (r *DatabaseRepository) Ping(ctx context.Context) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}

	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}
