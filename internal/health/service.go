package health

import (
	"context"
	"errors"
	"time"
)

type HealthService struct {
	repo    Repository
	appName string
	version string
}

func NewService(repo Repository, appName, version string) Service {
	return &HealthService{repo: repo, appName: appName, version: version}
}

func (s *HealthService) Check(ctx context.Context) Status {
	status := "pass"
	dbStatus := "connected"

	err := s.repo.Ping(ctx)
	if err != nil {
		if errors.Is(err, ErrDatabaseNotInitialized) {
			dbStatus = "not_initialized"
		} else {
			status = "fail"
			dbStatus = "disconnected"
		}
	}

	now := time.Now()
	return Status{
		Status:      status,
		Version:     s.version,
		ReleaseID:   now.Format("2006-01-02"),
		Description: s.appName + " API",
		DBStatus:    dbStatus,
		CheckedAt:   now.Format(time.RFC3339),
	}
}
