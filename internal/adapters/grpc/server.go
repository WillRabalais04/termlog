package grpc

import (
	"context"

	log "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/ports"
)

type Adapter struct {
	api ports.APIPort
	log.UnsafeLogServiceServer
}

func NewAdapter(api ports.APIPort) *Adapter {
	return &Adapter{api: api}
}

func (a *Adapter) Log(ctx context.Context, req *log.LogEntry) (*log.LogResponse, error) {
	entry := domain.LogEntry{
		Command:              req.Command,
		ExitCode:             req.ExitCode,
		Timestamp:            req.Timestamp,
		Shell_PID:            req.Shell_PID,
		ShellUptime:          req.ShellUptime,
		WorkingDirectory:     req.WorkingDirectory,
		PrevWorkingDirectory: req.PrevWorkingDirectory,
		User:                 req.User,
		EUID:                 req.EUID,
		Term:                 req.Term,
		Hostname:             req.Hostname,
		SSHClient:            req.SSHClient,
		TTY:                  req.TTY,
		IsGitRepo:            req.IsGitRepo,
		GitRepoRoot:          req.GitRepoRoot,
		GitBranch:            req.GitBranch,
		GitCommit:            req.GitCommit,
		GitStatus:            req.GitStatus,
		LoggedSuccessfully:   req.LoggedSuccessfully,
	}

	err := a.api.Log(entry)
	var successful bool
	if err == nil {
		successful = true
	}
	return &log.LogResponse{Success: successful}, err
}
