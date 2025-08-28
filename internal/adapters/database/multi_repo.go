package database

import (
	"context"
	"fmt"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type MultiRepo struct {
	cache  *LogRepo
	remote *LogRepo
}

func NewMultiRepo(cache, remote *LogRepo) *MultiRepo {
	return &MultiRepo{remote: remote, cache: cache}
}

func (r *MultiRepo) Log(ctx context.Context, entry *domain.LogEntry) error {
	if err := r.remote.Log(ctx, entry); err != nil {
		_ = r.cache.Log(ctx, entry)
		return fmt.Errorf("cached log due to remote repo failure: %w", err)
	}
	return nil
}

func (r *MultiRepo) Get(ctx context.Context, id int) (domain.LogEntry, error) {
	entry, err := r.remote.Get(ctx, id)
	if err != nil {
		return r.cache.Get(ctx, id)
	}
	return entry, nil
}

func (r *MultiRepo) List(ctx context.Context) ([]domain.LogEntry, error) {
	entries, err := r.remote.List(ctx)
	if err != nil {
		return r.cache.List(ctx)
	}
	return entries, nil
}
