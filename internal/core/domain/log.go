package domain

import (
	log "github.com/WillRabalais04/terminalLog/api/gen"
)

type LogEntry struct {
	EventID              string
	Command              string
	ExitCode             int32
	Timestamp            int64
	Shell_PID            int32
	ShellUptime          int64
	WorkingDirectory     string
	PrevWorkingDirectory string
	User                 string
	EUID                 int32
	Term                 string
	Hostname             string
	SSHClient            string
	TTY                  string
	IsGitRepo            bool
	GitRepoRoot          string
	GitBranch            string
	GitCommit            string
	GitStatus            string
	LoggedSuccessfully   bool
}

func ReqToDomainLogEntry(req *log.LogEntry) LogEntry {
	return LogEntry{
		EventID:              req.EventId,
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
}
