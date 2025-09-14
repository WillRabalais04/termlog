package grpc

import (
	"context"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
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
	domainEntry := LogEntryFromProto(req.Entry)
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
	return LogEntryToProto(entry), nil
}

func (a *Adapter) List(ctx context.Context, req *pb.ListLogsRequest) (*pb.ListLogsResponse, error) {
	filters := FilterFromProto(req.Filter)

	entries, err := a.api.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	protoEntries := make([]*pb.LogEntry, len(entries))
	for i, entry := range entries {
		protoEntries[i] = LogEntryToProto(entry)
	}

	return &pb.ListLogsResponse{Logs: protoEntries}, nil
}

func (a *Adapter) Delete(ctx context.Context, req *pb.DeleteLogRequest) (*pb.DeleteLogResponse, error) {
	deleted, err := a.api.Delete(ctx, req.GetEventId())
	if err != nil {
		return nil, err
	}
	return &pb.DeleteLogResponse{Success: true, Deleted: LogEntryToProto(deleted)}, nil
}

func (a *Adapter) DeleteMultipleLogs(ctx context.Context, req *pb.DeleteMultipleLogsRequest) (*pb.DeleteMultipleLogsResponse, error) {
	filters := FilterFromProto(req.Filter)

	deleted, err := a.api.DeleteMultiple(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteMultipleLogsResponse{Success: true, Deleted: LogEntriesToProto(deleted)}, nil
}
