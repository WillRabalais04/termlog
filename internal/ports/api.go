package ports

import (
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type APIPort interface {
	Log(entry domain.LogEntry) error
}
