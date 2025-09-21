package grpc

import (
	"context"
	"log"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
)

type ServerAdapter struct {
	svc *service.LogService
	pb.UnimplementedLogServiceServer
}

func NewServerAdapter(svc *service.LogService) *ServerAdapter {
	return &ServerAdapter{svc: svc}
}

func (a *ServerAdapter) Log(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	entries := LogEntriesFromProto(req.GetEntries())

	if len(entries) == 0 {
		log.Print("🔼 empty log request")
	} else if len(entries) == 1 {
		log.Printf("🔼 log request for log with command: '%s'", entries[0].Command)
	} else {
		log.Print("🔼 log request for logs with commands:")
		for _, entry := range entries {
			log.Printf("\t- %s", entry.Command)
		}
	}

	if err := a.svc.Log(ctx, entries); err != nil {
		log.Print("🔽 no entries logged")
		return nil, err
	}

	loggedEntryIDs := make([]string, 0, len(entries))
	if len(entries) > 0 {
		log.Print("🔽 logged entries with id's:")
		for _, entry := range entries {
			log.Printf("\t- %s", entry.EventID)
			loggedEntryIDs = append(loggedEntryIDs, entry.EventID)
		}
	}

	return &pb.LogResponse{Success: true, EventIds: loggedEntryIDs}, nil
}

func (a *ServerAdapter) Get(ctx context.Context, req *pb.GetRequest) (*pb.LogEntry, error) {
	eventID := req.GetEventId()
	log.Printf("🔼 get request for log (id: '%s')", eventID)

	entry, err := a.svc.Get(ctx, eventID)
	if err != nil {
		log.Printf("🔽 no log found for id: '%s'", eventID)
		return nil, err
	}

	log.Printf("🔽 found log (%s)", entry.EventID)
	return LogEntryToProto(entry), nil
}

func (a *ServerAdapter) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	filters := FilterFromProto(req.GetFilter())
	log.Printf("🔼 list request with filter: {%s}", FilterToString(filters))

	entries, err := a.svc.List(ctx, filters)
	if err != nil {
		log.Print("🔽 failed to list entries")
		return nil, err
	}

	if len(entries) == 0 {
		log.Print("🔽 no entries found matching filters")
	} else {
		log.Printf("🔽 found %d entries:", len(entries))
		for _, entry := range entries {
			if entry != nil {
				log.Printf("\t- %s", entry.EventID)
			}
		}
	}

	return &pb.ListResponse{Logs: LogEntriesToProto(entries)}, nil
}

func (a *ServerAdapter) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	eventID := req.GetEventId()
	log.Printf("🔼 delete request for log (id: '%s')", eventID)

	deleted, err := a.svc.Delete(ctx, eventID)
	if err != nil {
		log.Printf("🔽 log not deleted for id: '%s'", eventID)
		return nil, err
	}

	log.Printf("🔽 deleted log (id: '%s')", deleted.EventID)
	return &pb.DeleteResponse{Success: true, Deleted: LogEntryToProto(deleted)}, nil
}

func (a *ServerAdapter) DeleteMultiple(ctx context.Context, req *pb.DeleteMultipleRequest) (*pb.DeleteMultipleResponse, error) {
	filters := FilterFromProto(req.GetFilter())
	log.Printf("🔼 deletemultiple request with filter: {%s}", FilterToString(filters))

	deleted, err := a.svc.DeleteMultiple(ctx, filters)
	if err != nil {
		log.Printf("🔽 logs not deleted")
		return nil, err
	}

	if len(deleted) == 0 {
		log.Print("🔽 no entries found matching filters to delete")
	} else {
		log.Printf("🔽 deleted %d entries:", len(deleted))
		for _, entry := range deleted {
			if entry != nil {
				log.Printf("\t- %s", entry.EventID)
			}
		}
	}

	return &pb.DeleteMultipleResponse{Success: true, Deleted: LogEntriesToProto(deleted)}, nil
}
