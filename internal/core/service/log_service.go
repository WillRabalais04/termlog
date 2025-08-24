package service

import (
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/ports"
)

type Service struct {
	repo ports.LogRepositoryPort
}

func New(repo ports.LogRepositoryPort) *Service {
	return &Service{
		repo: repo,
	}
}
func (s *Service) Log(entry domain.LogEntry) error {
	return s.repo.Save(entry)
}
