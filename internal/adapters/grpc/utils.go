package grpc

import (
	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
)

func LogEntryToProto(entry *domain.LogEntry) *pb.LogEntry {
	return &pb.LogEntry{
		EventId:              entry.EventID,
		Command:              entry.Command,
		ExitCode:             entry.ExitCode,
		Timestamp:            entry.Timestamp,
		Shell_PID:            entry.Shell_PID,
		ShellUptime:          entry.ShellUptime,
		WorkingDirectory:     entry.WorkingDirectory,
		PrevWorkingDirectory: entry.PrevWorkingDirectory,
		User:                 entry.User,
		EUID:                 entry.EUID,
		Term:                 entry.Term,
		Hostname:             entry.Hostname,
		SSHClient:            entry.SSHClient,
		TTY:                  entry.TTY,
		GitRepo:              entry.GitRepo,
		GitRepoRoot:          entry.GitRepoRoot,
		GitBranch:            entry.GitBranch,
		GitCommit:            entry.GitCommit,
		GitStatus:            entry.GitStatus,
		LoggedSuccessfully:   entry.LoggedSuccessfully,
	}
}

func LogEntriesToProto(entries []*domain.LogEntry) []*pb.LogEntry {
	out := make([]*pb.LogEntry, 0, len(entries))

	for _, entry := range entries {
		out = append(out, LogEntryToProto(entry))
	}
	return out
}

func LogEntryFromProto(entry *pb.LogEntry) *domain.LogEntry {
	return &domain.LogEntry{
		EventID:              entry.EventId,
		Command:              entry.Command,
		ExitCode:             entry.ExitCode,
		Timestamp:            entry.Timestamp,
		Shell_PID:            entry.Shell_PID,
		ShellUptime:          entry.ShellUptime,
		WorkingDirectory:     entry.WorkingDirectory,
		PrevWorkingDirectory: entry.PrevWorkingDirectory,
		User:                 entry.User,
		EUID:                 entry.EUID,
		Term:                 entry.Term,
		Hostname:             entry.Hostname,
		SSHClient:            entry.SSHClient,
		TTY:                  entry.TTY,
		GitRepo:              entry.GitRepo,
		GitRepoRoot:          entry.GitRepoRoot,
		GitBranch:            entry.GitBranch,
		GitCommit:            entry.GitCommit,
		GitStatus:            entry.GitStatus,
		LoggedSuccessfully:   entry.LoggedSuccessfully,
	}
}

// func LogEntriesFromProto(entries []*pb.LogEntry) []*LogEntry {
// 	out := make([]*LogEntry, 0, len(entries))

// 	for _, entry := range entries {
// 		out = append(out, logEntryFromProto(entry))
// 	}
// 	return out
// }

func FilterFromProto(filter *pb.LogFilter) *domain.LogFilter {
	if filter == nil {
		return &domain.LogFilter{}
	}

	portsFilterTerms := make(map[string]domain.FilterValues)
	for key, values := range filter.FilterTerms {
		portsFilterTerms[key] = domain.FilterValues{
			Values: values.Values,
		}
	}

	portsSearchTerms := make(map[string]domain.SearchValues)
	for key, values := range filter.SearchTerms {
		portsSearchTerms[key] = domain.SearchValues{
			Values: values.Values,
		}
	}
	return &domain.LogFilter{ // review type selection
		FilterTerms: portsFilterTerms,
		FilterMode:  domain.Mode(*filter.FilterMode),
		SearchTerms: portsSearchTerms,
		SearchMode:  domain.Mode(*filter.SearchMode),
		Limit:       *filter.Limit,
		Offset:      *filter.Offset,
		OrderBy:     filter.OrderBy,
		StartTime:   &filter.StartTime.Seconds,
		EndTime:     &filter.EndTime.Seconds,
	}
}
