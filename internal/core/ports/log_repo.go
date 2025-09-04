package ports

import (
	"context"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

type LogRepositoryPort interface {
	Log(ctx context.Context, entry *domain.LogEntry) error
	Get(ctx context.Context, id string) (*domain.LogEntry, error)
	List(ctx context.Context, filters *LogQuery) ([]*domain.LogEntry, error)
	Delete(ctx context.Context, id string) error
	DeleteMultiple(ctx context.Context, filters *LogQuery) error
}

type LogQuery struct {
	EventID              *string // verify accessing by pointer is best practices but works for nil check
	Command              *string
	User                 *string
	ExitCode             *int
	Limit                uint64
	Offset               uint64
	OrderBy              *string
	GitRepo              *bool
	Timestamp            *int64
	ShellPID             *uint64
	ShellUptime          *uint64
	WorkingDirectory     *string
	PrevWorkingDirectory *string
	EUID                 *uint64 // make uint32?
	Term                 *string
	Hostname             *string
	SSHClient            *string
	TTY                  *string
	GitRepoRoot          *string
	Branch               *string
	Commit               *string
	LoggedSuccesfully    *bool
}
