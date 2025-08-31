package domain

import (
	"log"

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

func PrintLogEntry(entry *LogEntry) { // handle nil cases

	log.Printf("LogEntry: {\n")
	log.Printf("EventID:		%s\n", entry.EventID)
	log.Printf("Command:		%s\n", entry.Command)
	log.Printf("ExitCode:		%d\n", entry.ExitCode)
	log.Printf("Timestamp:		%d\n", entry.Timestamp)
	log.Printf("Shell_PID:		%d\n", entry.Shell_PID)
	log.Printf("ShellUptime:	%d\n", entry.ShellUptime)
	log.Printf("WorkingDirectory:		%s\n", entry.WorkingDirectory)
	log.Printf("PrevWorkingDirectory:		%s\n", entry.PrevWorkingDirectory)
	log.Printf("User:		%s\n", entry.User)
	log.Printf("EUID:		%d\n", entry.EUID)
	log.Printf("Term:		%s\n", entry.Term)
	log.Printf("Hostname:		%s\n", entry.Hostname)
	log.Printf("TTY:		%s\n", entry.TTY)
	log.Printf("IsGitRepo:		%t\n", entry.IsGitRepo)
	log.Printf("GitRepoRoot:		%s\n", entry.GitRepoRoot)
	log.Printf("GitBranch:		%s\n", entry.GitBranch)
	log.Printf("GitCommit:		%s\n", entry.GitCommit)
	log.Printf("GitStatus:		%s\n", entry.GitStatus)
	log.Printf("LoggedSuccesfully:		%t\n", entry.LoggedSuccessfully)
	log.Println("}")
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
