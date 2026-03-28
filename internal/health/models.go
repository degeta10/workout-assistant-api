package health

import (
	"context"
	"errors"
)

var ErrDatabaseNotInitialized = errors.New("database not initialized")

type Status struct {
	Status      string
	Version     string
	ReleaseID   string
	Description string
	DBStatus    string
	CheckedAt   string
}

type Repository interface {
	Ping(ctx context.Context) error
}

type Service interface {
	Check(ctx context.Context) Status
}
