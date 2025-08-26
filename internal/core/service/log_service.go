package service

import (
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
func (s *LogService) Log(entry domain.LogEntry) error {
	return s.repo.Save(entry)
}
