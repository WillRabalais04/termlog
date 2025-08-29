package service

import (
	"context"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/ports"
)

type LogService struct {
	repo ports.LogRepositoryPort
}

func NewLogService(repo ports.LogRepositoryPort) *LogService {
	return &LogService{
		repo: repo,
	}
}
func (s *LogService) Log(ctx context.Context, entry *domain.LogEntry) error {
	return s.repo.Log(ctx, entry)
}
func (s *LogService) Get(ctx context.Context, id int) (domain.LogEntry, error) {
	return s.repo.Get(ctx, id)
}
func (s *LogService) List(ctx context.Context) ([]domain.LogEntry, error) {
	return s.repo.List(ctx)
}
