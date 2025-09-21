package grpc

import (
	"fmt"
	"strings"
	"time"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		EventID:              entry.GetEventId(),
		Command:              entry.GetCommand(),
		ExitCode:             entry.GetExitCode(),
		Timestamp:            entry.GetTimestamp(),
		Shell_PID:            entry.GetShell_PID(),
		ShellUptime:          entry.GetShellUptime(),
		WorkingDirectory:     entry.GetWorkingDirectory(),
		PrevWorkingDirectory: entry.GetPrevWorkingDirectory(),
		User:                 entry.GetUser(),
		EUID:                 entry.GetEUID(),
		Term:                 entry.GetTerm(),
		Hostname:             entry.GetHostname(),
		SSHClient:            entry.GetSSHClient(),
		TTY:                  entry.GetTTY(),
		GitRepo:              entry.GetGitRepo(),
		GitRepoRoot:          entry.GetGitRepoRoot(),
		GitBranch:            entry.GetGitBranch(),
		GitCommit:            entry.GetGitCommit(),
		GitStatus:            entry.GetGitStatus(),
		LoggedSuccessfully:   entry.GetLoggedSuccessfully(),
	}
}

func LogEntriesFromProto(entries []*pb.LogEntry) []*domain.LogEntry {
	out := make([]*domain.LogEntry, 0, len(entries))

	for _, entry := range entries {
		out = append(out, LogEntryFromProto(entry))
	}
	return out
}

func FilterToProto(filter *domain.LogFilter) *pb.LogFilter {
	if filter == nil {
		return &pb.LogFilter{}
	}

	protoFilter := &pb.LogFilter{
		FilterTerms: make(map[string]*pb.FilterValues),
		SearchTerms: make(map[string]*pb.SearchValues),
		OrderBy:     filter.OrderBy,
	}

	for key, values := range filter.FilterTerms {
		protoFilter.FilterTerms[key] = &pb.FilterValues{
			Values: values.Values,
		}
	}
	for key, values := range filter.SearchTerms {
		protoFilter.SearchTerms[key] = &pb.SearchValues{
			Values: values.Values,
		}
	}

	filterMode := pb.FilterMode(int32(filter.FilterMode))
	protoFilter.FilterMode = &filterMode

	searchMode := pb.SearchMode(int32(filter.SearchMode))
	protoFilter.SearchMode = &searchMode

	if filter.StartTime != nil {
		protoFilter.StartTime = timestamppb.New(time.Unix(*filter.StartTime, 0))
	}

	if filter.EndTime != nil {
		protoFilter.EndTime = timestamppb.New(time.Unix(*filter.EndTime, 0))
	}

	if filter.Limit > 0 {
		protoFilter.Limit = &filter.Limit
	}

	if filter.Offset > 0 {
		protoFilter.Offset = &filter.Offset
	}

	return protoFilter
}

func FilterFromProto(protoFilter *pb.LogFilter) *domain.LogFilter {
	if protoFilter == nil {
		return &domain.LogFilter{}
	}

	domainFilter := &domain.LogFilter{
		FilterTerms: make(map[string]domain.FilterValues),
		SearchTerms: make(map[string]domain.SearchValues),
	}

	for key, values := range protoFilter.GetFilterTerms() {
		if values != nil {
			domainFilter.FilterTerms[key] = domain.FilterValues{
				Values: values.Values,
			}
		}
	}
	for key, values := range protoFilter.SearchTerms {
		if values != nil {
			domainFilter.SearchTerms[key] = domain.SearchValues{
				Values: values.Values,
			}
		}
	}

	if protoFilter.FilterMode != nil {
		domainFilter.FilterMode = domain.Mode(*protoFilter.FilterMode)
	}
	if protoFilter.SearchMode != nil {
		domainFilter.SearchMode = domain.Mode(*protoFilter.SearchMode)
	}
	if protoFilter.Limit != nil {
		domainFilter.Limit = *protoFilter.Limit
	}
	if protoFilter.Offset != nil {
		domainFilter.Offset = *protoFilter.Offset
	}

	domainFilter.OrderBy = protoFilter.OrderBy

	if protoFilter.StartTime != nil {
		startTime := protoFilter.StartTime.AsTime().Unix()
		domainFilter.StartTime = &startTime
	}
	if protoFilter.EndTime != nil {
		endTime := protoFilter.EndTime.AsTime().Unix()
		domainFilter.EndTime = &endTime
	}
	return domainFilter
}

func FilterToString(filter *domain.LogFilter) string {
	if filter == nil {
		return "filter: (nil)"
	}

	var parts []string

	if len(filter.FilterTerms) > 0 {
		var termParts []string
		for key, values := range filter.FilterTerms {
			termParts = append(termParts, fmt.Sprintf("%s=[%s]", key, strings.Join(values.Values, ",")))
		}
		parts = append(parts, fmt.Sprintf("filters: {%s}", strings.Join(termParts, " ")))
	}
	if len(filter.SearchTerms) > 0 {
		var termParts []string
		for key, values := range filter.SearchTerms {
			termParts = append(termParts, fmt.Sprintf("%s~=[%s]", key, strings.Join(values.Values, ",")))
		}
		parts = append(parts, fmt.Sprintf("searches: {%s}", strings.Join(termParts, " ")))
	}
	if filter.StartTime != nil {
		parts = append(parts, fmt.Sprintf("startTime: %s", time.Unix(*filter.StartTime, 0).Format(time.RFC3339)))
	}
	if filter.EndTime != nil {
		parts = append(parts, fmt.Sprintf("endTime: %s", time.Unix(*filter.EndTime, 0).Format(time.RFC3339)))
	}
	if filter.OrderBy != nil {
		parts = append(parts, fmt.Sprintf("ordering: %q", *filter.OrderBy))
	}
	if filter.Limit > 0 {
		parts = append(parts, fmt.Sprintf("limit: %d", filter.Limit))
	}
	if filter.Offset > 0 {
		parts = append(parts, fmt.Sprintf("offset: %d", filter.Offset))
	}
	if len(parts) == 0 {
		return "filter: (empty)"
	}

	return strings.Join(parts, " | ")
}
