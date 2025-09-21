package grpc

import (
	"context"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/ports"
	"google.golang.org/grpc"
)

type ClientAdapter struct {
	client pb.LogServiceClient
}

func NewClientAdapter(conn *grpc.ClientConn) ports.LogRepositoryPort {
	return &ClientAdapter{client: pb.NewLogServiceClient(conn)}
}

func (c *ClientAdapter) Log(ctx context.Context, entries []*domain.LogEntry) error {
	_, err := c.client.Log(ctx, &pb.LogRequest{
		Entries: LogEntriesToProto(entries),
	})
	return err
}

func (c *ClientAdapter) Get(ctx context.Context, id string) (*domain.LogEntry, error) {
	resp, err := c.client.Get(ctx, &pb.GetRequest{EventId: id})
	if err != nil {
		return nil, err
	}
	return LogEntryFromProto(resp), nil
}

func (c *ClientAdapter) List(ctx context.Context, filter *domain.LogFilter) ([]*domain.LogEntry, error) {
	resp, err := c.client.List(ctx, &pb.ListRequest{
		Filter: FilterToProto(filter),
	})
	if err != nil {
		return nil, err
	}
	return LogEntriesFromProto(resp.Logs), nil
}

func (c *ClientAdapter) Delete(ctx context.Context, id string) (*domain.LogEntry, error) {
	resp, err := c.client.Delete(ctx, &pb.DeleteRequest{EventId: id})
	if err != nil {
		return nil, err
	}
	return LogEntryFromProto(resp.Deleted), nil
}

func (c *ClientAdapter) DeleteMultiple(ctx context.Context, filter *domain.LogFilter) ([]*domain.LogEntry, error) {
	resp, err := c.client.DeleteMultiple(ctx, &pb.DeleteMultipleRequest{
		Filter: FilterToProto(filter),
	})
	if err != nil {
		return nil, err
	}
	return LogEntriesFromProto(resp.Deleted), nil
}
