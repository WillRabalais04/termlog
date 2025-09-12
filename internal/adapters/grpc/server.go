package grpc

import (
	"context"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/ports"
)

type Adapter struct {
	api ports.APIPort
	pb.UnimplementedLogServiceServer
}

func NewAdapter(api ports.APIPort) *Adapter {
	return &Adapter{api: api}
}

func (a *Adapter) Log(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	domainEntry := logEntryFromProto(req.Entry)
	err := a.api.Log(ctx, domainEntry)
	if err != nil {
		return nil, err
	}
	return &pb.LogResponse{Success: true, EventId: domainEntry.EventID}, nil
}

func (a *Adapter) GetLog(ctx context.Context, req *pb.GetLogRequest) (*pb.LogEntry, error) {
	entry, err := a.api.Get(ctx, req.GetEventId())
	if err != nil {
		return nil, err
	}
	return logEntryToProto(entry), nil
}

func (a *Adapter) List(ctx context.Context, req *pb.ListLogsRequest) (*pb.ListLogsResponse, error) {
	filters := filterFromProto(req.Filter)

	entries, err := a.api.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	protoEntries := make([]*pb.LogEntry, len(entries))
	for i, entry := range entries {
		protoEntries[i] = logEntryToProto(entry)
	}

	return &pb.ListLogsResponse{Logs: protoEntries}, nil
}

func (a *Adapter) Delete(ctx context.Context, req *pb.DeleteLogRequest) (*pb.DeleteLogResponse, error) {
	err := a.api.Delete(ctx, req.GetEventId())
	if err != nil {
		return nil, err
	}
	return &pb.DeleteLogResponse{Success: true}, nil
}

func (a *Adapter) DeleteMultipleLogs(ctx context.Context, req *pb.DeleteMultipleLogsRequest) (*pb.DeleteMultipleLogsResponse, error) {
	filters := filterFromProto(req.Filter)

	err := a.api.DeleteMultiple(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteMultipleLogsResponse{Success: true, DeletedCount: 0}, nil
}

func logEntryToProto(entry *domain.LogEntry) *pb.LogEntry {
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
		IsGitRepo:            entry.IsGitRepo,
		GitRepoRoot:          entry.GitRepoRoot,
		GitBranch:            entry.GitBranch,
		GitCommit:            entry.GitCommit,
		GitStatus:            entry.GitStatus,
		LoggedSuccessfully:   entry.LoggedSuccessfully,
	}
}

func logEntryFromProto(entry *pb.LogEntry) *domain.LogEntry {
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
		IsGitRepo:            entry.IsGitRepo,
		GitRepoRoot:          entry.GitRepoRoot,
		GitBranch:            entry.GitBranch,
		GitCommit:            entry.GitCommit,
		GitStatus:            entry.GitStatus,
		LoggedSuccessfully:   entry.LoggedSuccessfully,
	}
}

func filterFromProto(filter *pb.LogFilter) *ports.LogFilter {
	if filter == nil {
		return &ports.LogFilter{}
	}

	portsFilterTerms := make(map[string]ports.FilterValues)
	for key, values := range filter.FilterTerms {
		portsFilterTerms[key] = ports.FilterValues{
			Values: values.Values,
		}
	}

	portsSearchTerms := make(map[string]ports.SearchValues)
	for key, values := range filter.SearchTerms {
		portsSearchTerms[key] = ports.SearchValues{
			Values: values.Values,
		}
	}
	return &ports.LogFilter{ // review type selection
		FilterTerms: portsFilterTerms,
		FilterMode:  ports.Mode(*filter.FilterMode),
		SearchTerms: portsSearchTerms,
		SearchMode:  ports.Mode(*filter.SearchMode),
		Limit:       *filter.Limit,
		Offset:      *filter.Offset,
		OrderBy:     filter.OrderBy,
		StartTime:   &filter.StartTime.Seconds,
		EndTime:     &filter.EndTime.Seconds,
	}
}
