package ports

import (
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type LogRepositoryPort interface {
	Save(entry domain.LogEntry) error
	Get(id int) (domain.LogEntry, error)
	List() ([]domain.LogEntry, error)
	// add delete and clear functions
}
