package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/ports"
)

type MultiRepo struct {
	cache  ports.LogRepositoryPort
	remote ports.LogRepositoryPort
}

func NewMultiRepo(cache, remote ports.LogRepositoryPort) *MultiRepo {
	return &MultiRepo{cache: cache, remote: remote}
}

func (r *MultiRepo) GetCache() ports.LogRepositoryPort {
	return r.cache
}
func (r *MultiRepo) GetRemote() ports.LogRepositoryPort {
	return r.remote
}

func (r *MultiRepo) Log(ctx context.Context, entries []*domain.LogEntry) error {
	if r.remote != nil {
		err := r.remote.Log(ctx, entries)
		if err == nil {
			return nil
		}
		log.Printf("remote log failed, falling back to cache: %v", err)
	}
	return r.cache.Log(ctx, entries)
}

func (r *MultiRepo) FlushCache(ctx context.Context) ([]*domain.LogEntry, error) {
	if r.remote == nil {
		return nil, nil
	}

	entries, err := r.cache.List(ctx, &domain.LogFilter{})
	if err != nil {
		return nil, fmt.Errorf("reading cache failed: %w", err)
	}
	if len(entries) == 0 {
		return nil, nil
	}

	if err := r.remote.Log(ctx, entries); err != nil {
		return nil, fmt.Errorf("failed to push cache entries to remote: %w", err)
	}

	succeededIDs := make([]string, len(entries))
	for i, entry := range entries {
		succeededIDs[i] = entry.EventID
	}

	filter := domain.NewFilterBuilder().SetFilterMode(domain.OR)
	for _, id := range succeededIDs {
		filter.AddFilterTerm("event_id", id)
	}

	_, err = r.cache.DeleteMultiple(ctx, filter.Build())
	if err != nil {
		log.Printf("CRITICAL: failed to delete flushed entries from cache: %v", err)
		return nil, err
	}

	log.Printf("successfully flushed %d entries.", len(entries))
	return entries, nil
}

func (r *MultiRepo) StartCacheFlusher(ctx context.Context, interval time.Duration, quit <-chan struct{}) {
	log.Println("starting background cache flusher...")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	r.FlushCache(ctx)

	for {
		select {
		case <-ticker.C:
			log.Println("checking for cache entries to flush...")
			if _, err := r.FlushCache(ctx); err != nil {
				log.Printf("background flush failed: %v", err)
			}
		case <-quit:
			log.Println("stopping cache flusher.")
			r.FlushCache(ctx)
			return
		}
	}
}

func (r *MultiRepo) Get(ctx context.Context, id string) (*domain.LogEntry, error) {
	entry, err := r.remote.Get(ctx, id)
	if err != nil {
		return r.cache.Get(ctx, id)
	}
	return entry, nil
}

func (r *MultiRepo) List(ctx context.Context, filters *domain.LogFilter) ([]*domain.LogEntry, error) {
	entries, err := r.remote.List(ctx, filters)
	if err != nil {
		return r.cache.List(ctx, filters)
	}
	return entries, nil
}

func (r *MultiRepo) Delete(ctx context.Context, id string) (*domain.LogEntry, error) {
	deleted, err1 := r.remote.Delete(ctx, id)
	if err1 != nil || deleted == nil {
		deleted, err2 := r.cache.Delete(ctx, id)
		if err2 != nil {
			return nil, fmt.Errorf("could not access local db: %v", err2)
		}
		return deleted, nil
	}
	return deleted, nil
}

func (r *MultiRepo) DeleteMultiple(ctx context.Context, filters *domain.LogFilter) ([]*domain.LogEntry, error) {
	deleted, err1 := r.remote.DeleteMultiple(ctx, filters)
	if err1 != nil || len(deleted) == 0 {
		deleted, err2 := r.cache.DeleteMultiple(ctx, filters)
		if err2 != nil {
			return nil, fmt.Errorf("delete multiple failed: could not access local or remote db: %v %v", err1, err2)
		}
		return deleted, nil
	}
	return deleted, nil
}
