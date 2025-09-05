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

	query := &ports.LogFilter{}

	if filter.EventId != nil {
		query.EventID = filter.EventId
	}
	if filter.Command != nil {
		query.Command = filter.Command
	}
	if filter.User != nil {
		query.User = filter.User
	}
	temp := int(*filter.ExitCode) // clean
	if filter.ExitCode != nil {
		query.ExitCode = &temp
	}
	if filter.IsGitRepo != nil {
		query.GitRepo = filter.IsGitRepo
	}
	temp2 := uint64(*filter.Shell_PID) // clean
	if filter.Shell_PID != nil {
		query.ShellPID = &temp2
	}
	temp2 = uint64(*filter.Shell_PID) // clean
	if filter.EUID != nil {
		query.EUID = &temp2
	}
	if filter.WorkingDirectory != nil {
		query.WorkingDirectory = filter.WorkingDirectory
	}
	if filter.Term != nil {
		query.Term = filter.Term
	}
	if filter.Hostname != nil {
		query.Hostname = filter.Hostname
	}
	if filter.GitRepoRoot != nil {
		query.GitRepoRoot = filter.GitRepoRoot
	}
	if filter.GitBranch != nil {
		query.Branch = filter.GitBranch
	}
	if filter.LoggedSuccessfully != nil {
		query.LoggedSuccessfully = filter.LoggedSuccessfully
	}

	if filter.StartTime != nil && filter.StartTime.IsValid() {
		startTime := filter.StartTime.GetSeconds()
		query.StartTime = &startTime
	}
	if filter.EndTime != nil && filter.EndTime.IsValid() {
		endTime := filter.EndTime.GetSeconds()
		query.EndTime = &endTime
	}

	return query
}
