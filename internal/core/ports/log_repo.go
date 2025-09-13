package ports

import (
	"context"
	"time"

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

type FilterBuilder struct {
	filter *LogFilter
}

func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		filter: &LogFilter{
			FilterTerms: make(map[string]FilterValues),
			SearchTerms: make(map[string]SearchValues),
			FilterMode:  OR,
			SearchMode:  OR,
		},
	}
}

func (b *FilterBuilder) AddFilterTerm(key string, values ...string) *FilterBuilder {
	b.filter.FilterTerms[key] = FilterValues{Values: values}
	return b
}

func (b *FilterBuilder) AddSearchTerm(key string, values ...string) *FilterBuilder {
	b.filter.SearchTerms[key] = SearchValues{Values: values}
	return b
}

func (b *FilterBuilder) SetFilterMode(mode Mode) *FilterBuilder {
	b.filter.FilterMode = mode
	return b
}

func (b *FilterBuilder) SetSearchMode(mode Mode) *FilterBuilder {
	b.filter.SearchMode = mode
	return b
}

func (b *FilterBuilder) SetLimit(limit uint64) *FilterBuilder {
	b.filter.Limit = limit
	return b
}

func (b *FilterBuilder) SetOffset(offset uint64) *FilterBuilder {
	b.filter.Offset = offset
	return b
}

func (b *FilterBuilder) SetOrderBy(orderBy string) *FilterBuilder {
	b.filter.OrderBy = &orderBy
	return b
}

func (b *FilterBuilder) SetTimeRange(start, end time.Time) *FilterBuilder {
	startTime := start.Unix()
	endTime := end.Unix()
	b.filter.StartTime = &startTime
	b.filter.EndTime = &endTime
	return b
}

func (b *FilterBuilder) Build() *LogFilter {
	return b.filter
}

type LogRepositoryPort interface {
	Log(ctx context.Context, entry *domain.LogEntry) error
	Get(ctx context.Context, id string) (*domain.LogEntry, error)
	List(ctx context.Context, filters *LogFilter) ([]*domain.LogEntry, error)
	Delete(ctx context.Context, id string) (*domain.LogEntry, error)
	DeleteMultiple(ctx context.Context, filters *LogFilter) ([]*domain.LogEntry, error)
}
