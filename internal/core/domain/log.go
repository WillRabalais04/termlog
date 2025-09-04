package domain

import (
	"fmt"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
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

func PrintLogEntry(entry *LogEntry) {

	fmt.Printf("LogEntry: {\n")
	fmt.Printf("EventID:		%s\n", entry.EventID)
	fmt.Printf("Command:		%s\n", entry.Command)
	fmt.Printf("ExitCode:		%d\n", entry.ExitCode)
	fmt.Printf("Timestamp:		%d\n", entry.Timestamp)
	fmt.Printf("Shell_PID:		%d\n", entry.Shell_PID)
	fmt.Printf("ShellUptime:	%d\n", entry.ShellUptime)
	fmt.Printf("WorkingDirectory:		%s\n", entry.WorkingDirectory)
	fmt.Printf("PrevWorkingDirectory:		%s\n", entry.PrevWorkingDirectory)
	fmt.Printf("User:		%s\n", entry.User)
	fmt.Printf("EUID:		%d\n", entry.EUID)
	fmt.Printf("Term:		%s\n", entry.Term)
	fmt.Printf("Hostname:		%s\n", entry.Hostname)
	fmt.Printf("TTY:		%s\n", entry.TTY)
	fmt.Printf("IsGitRepo:		%t\n", entry.IsGitRepo)
	fmt.Printf("GitRepoRoot:		%s\n", entry.GitRepoRoot)
	fmt.Printf("GitBranch:		%s\n", entry.GitBranch)
	fmt.Printf("GitCommit:		%s\n", entry.GitCommit)
	fmt.Printf("GitStatus:		%s\n", entry.GitStatus)
	fmt.Printf("LoggedSuccesfully:		%t\n", entry.LoggedSuccessfully)
	fmt.Println("}")
}

func ReqToDomainLogEntry(req *pb.LogEntry) LogEntry {
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
