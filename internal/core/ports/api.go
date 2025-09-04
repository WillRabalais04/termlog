package ports

import (
	"context"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type APIPort interface {
	Log(ctx context.Context, entry *domain.LogEntry) error
	Get(ctx context.Context, id string) (*domain.LogEntry, error)
	List(ctx context.Context, filters *LogQuery) ([]*domain.LogEntry, error)
	Delete(ctx context.Context, id string) error
	DeleteMultiple(ctx context.Context, filters *LogQuery) error
}
