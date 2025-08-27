package memory

import (
	"fmt"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type APIPort interface {
	Log(entry *domain.LogEntry) error
}

type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Save(entry domain.LogEntry) error {
	fmt.Printf("--- LOG SAVED ---\n")
	fmt.Printf("User: %s, Command: %s\n", entry.User, entry.Command)
	fmt.Printf("--------------------\n")
	return nil
}
