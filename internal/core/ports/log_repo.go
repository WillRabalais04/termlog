package ports

import (
	"context"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type LogRepositoryPort interface {
	Log(ctx context.Context, entry *domain.LogEntry) error
	Get(ctx context.Context, id int) (domain.LogEntry, error)
	List(ctx context.Context) ([]domain.LogEntry, error)
	// add delete and clear functions
}


