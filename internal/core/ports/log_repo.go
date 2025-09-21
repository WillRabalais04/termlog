package ports

import (
	"context"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type LogRepositoryPort interface {
	Log(ctx context.Context, entry []*domain.LogEntry) error
	Get(ctx context.Context, id string) (*domain.LogEntry, error)
	List(ctx context.Context, filters *domain.LogFilter) ([]*domain.LogEntry, error)
	Delete(ctx context.Context, id string) (*domain.LogEntry, error) // probably should refactor into just one delete
	DeleteMultiple(ctx context.Context, filters *domain.LogFilter) ([]*domain.LogEntry, error)
}
