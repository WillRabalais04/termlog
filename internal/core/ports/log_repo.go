package ports

import (
	"context"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type FilterValues struct {
	Values []string
}

type SearchValues struct {
	Values []string
}

type Mode int

const (
	OR  Mode = 0
	AND Mode = 1
)

type LogFilter struct {
	FilterTerms map[string]FilterValues
	FilterMode  Mode
	SearchTerms map[string]SearchValues
	SearchMode  Mode
	Limit       uint64
	Offset      uint64
	OrderBy     *string
	StartTime   *int64
	EndTime     *int64
}

type LogRepositoryPort interface {
	Log(ctx context.Context, entry *domain.LogEntry) error
	Get(ctx context.Context, id string) (*domain.LogEntry, error)
	List(ctx context.Context, filters *LogFilter) ([]*domain.LogEntry, error)
	Delete(ctx context.Context, id string) error
	DeleteMultiple(ctx context.Context, filters *LogFilter) error
}
