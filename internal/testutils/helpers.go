package testutils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/google/uuid"
)

func LogEntryToString(entry *domain.LogEntry) string {
	if entry == nil {
		return "[nil log entry]"
	}
	var s string
	s += "LogEntry: {\n"
	s += fmt.Sprintf("EventID:		%s\n", entry.EventID)
	s += fmt.Sprintf("Command:		%s\n", entry.Command)
	s += fmt.Sprintf("ExitCode:		%d\n", entry.ExitCode)
	s += fmt.Sprintf("Timestamp:		%d\n", entry.Timestamp)
	s += fmt.Sprintf("Shell_PID:		%d\n", entry.Shell_PID)
	s += fmt.Sprintf("ShellUptime:	%d\n", entry.ShellUptime)
	s += fmt.Sprintf("WorkingDirectory:		%s\n", entry.WorkingDirectory)
	s += fmt.Sprintf("PrevWorkingDirectory:		%s\n", entry.PrevWorkingDirectory)
	s += fmt.Sprintf("User:		%s\n", entry.User)
	s += fmt.Sprintf("EUID:		%d\n", entry.EUID)
	s += fmt.Sprintf("Term:		%s\n", entry.Term)
	s += fmt.Sprintf("Hostname:		%s\n", entry.Hostname)
	s += fmt.Sprintf("TTY:		%s\n", entry.TTY)
	s += fmt.Sprintf("GitRepo:		%t\n", entry.GitRepo)
	s += fmt.Sprintf("GitRepoRoot:		%s\n", entry.GitRepoRoot)
	s += fmt.Sprintf("GitBranch:		%s\n", entry.GitBranch)
	s += fmt.Sprintf("GitCommit:		%s\n", entry.GitCommit)
	s += fmt.Sprintf("GitStatus:		%s\n", entry.GitStatus)
	s += fmt.Sprintf("LoggedSuccessfully:		%t\n", entry.LoggedSuccessfully)
	s += "}\n"

	return s
}
func LogEntriesToString(entries []*domain.LogEntry) string {

	var s string

	for _, entry := range entries {
		s += LogEntryToString(entry) + "\n"
	}

	return s
}
func PrettyPrintString(testName string, entries []*domain.LogEntry, err error) string {
	var s string
	width := 80
	label := fmt.Sprintf(" Test: %s ", testName)

	left := (width - len(label)) / 2
	right := width - len(label) - left

	s += fmt.Sprintf("%s%s%s\n", strings.Repeat("-", left), label, strings.Repeat("-", right))

	if err != nil {
		s += fmt.Sprintf("%v\n", err)
	} else {
		for _, entry := range entries {
			s += LogEntryToString(entry)
		}
	}
	s += strings.Repeat("-", width) + "\n"

	return s
}

func PrettyPrint(testName string, entries []*domain.LogEntry, err error) {
	fmt.Print(PrettyPrintString(testName, entries, err))
}

func RandomLog() *domain.LogEntry { // commented out fields because the printed outputs became unreadable
	return &domain.LogEntry{
		Command:   fmt.Sprintf("cmd-%d", rand.Intn(100)),
		ExitCode:  rand.Int31n(10),
		Timestamp: time.Now().UnixNano(),
		Shell_PID: int32(rand.Intn(5000)),
		// ShellUptime:          rand.Int63n(1000000),
		WorkingDirectory: fmt.Sprintf("/tmp/dir-%d", rand.Intn(5)),
		// PrevWorkingDirectory: "/" + uuid.New().String()[:8],
		User: "user" + uuid.New().String()[:8],
		EUID: int32(rand.Intn(1000)),
		// Term:                 uuid.New().String()[:8],
		// Hostname:             uuid.New().String()[:8],
		// SSHClient:            "",
		// TTY:                  fmt.Sprintf("pts/%d", rand.Intn(10)),
		GitRepo: rand.Intn(2) == 0,
		// GitRepoRoot:          "/" + uuid.New().String()[:8],
		// GitBranch:            uuid.New().String()[:8],
		// GitCommit:            uuid.New().String()[:8],
		// GitStatus:            "clean",
		// LoggedSuccessfully:   true,
	}
}
