package database

import (
	"context"
	"fmt"
	"log"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/ports"
)

type MultiRepo struct {
	cache  *LogRepo
	remote *LogRepo
}

func NewMultiRepo(cache, remote *LogRepo) *MultiRepo {
	return &MultiRepo{cache: cache, remote: remote}
}

func (r *MultiRepo) Log(ctx context.Context, entry *domain.LogEntry) error {
	if err := r.flushCache(ctx); err != nil {
		log.Printf("warning: failed to flush cache: %v\n", err)
	}

	if err := r.remote.Log(ctx, entry); err != nil {
		if errCache := r.cache.Log(ctx, entry); errCache != nil {
			return fmt.Errorf("failed remote (%v) and cache (%v)", err, errCache)
		}
		return fmt.Errorf("remote failed, cached entry: %w", err)
	}

	return nil
}

func (r *MultiRepo) flushCache(ctx context.Context) error {
	entries, err := r.cache.List(ctx, &ports.LogFilter{}) // empty filter should return all

	if err != nil {
		return fmt.Errorf("reading cache failed: %w", err)
	}
	if len(entries) == 0 {
		return nil
	}

	for _, e := range entries {
		if err := r.remote.Log(ctx, e); err != nil {
			return fmt.Errorf("failed to push cached entry %s: %w", e.EventID, err)
		}
		if _, err := r.cache.Delete(ctx, e.EventID); err != nil {
			return fmt.Errorf("failed to delete cached entry %s: %w", e.EventID, err)
		}
	}
	return nil
}

func (r *MultiRepo) Get(ctx context.Context, id string) (*domain.LogEntry, error) {
	entry, err := r.remote.Get(ctx, id)
	if err != nil {
		return r.cache.Get(ctx, id)
	}
	return entry, nil
}

func (r *MultiRepo) List(ctx context.Context, filters *ports.LogFilter) ([]*domain.LogEntry, error) {
	entries, err := r.remote.List(ctx, filters)
	if err != nil {
		return r.cache.List(ctx, filters)
	}
	return entries, nil
}

func (r *MultiRepo) Delete(ctx context.Context, id string) (*domain.LogEntry, error) {
	deleted, err1 := r.remote.Delete(ctx, id)
	if err1 != nil {
		deleted, err2 := r.cache.Delete(ctx, id)
		if err2 != nil {
			return nil, fmt.Errorf("could not access local db: %v", err2)
		}
		return deleted, nil
	}
	return deleted, nil
}

func (r *MultiRepo) DeleteMultiple(ctx context.Context, filters *ports.LogFilter) ([]*domain.LogEntry, error) {
	deleted, err1 := r.remote.DeleteMultiple(ctx, filters)
	if err1 != nil {
		deleted, err2 := r.cache.DeleteMultiple(ctx, filters)
		if err2 != nil {
			return nil, fmt.Errorf("delete multiple failed: could not access local or remote db: %v %v", err1, err2)
		}
		return deleted, nil
	}
	return deleted, nil
}
