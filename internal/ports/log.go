package ports

import (
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type LogRepositoryPort interface {
	Save(entry domain.LogEntry) error
}
