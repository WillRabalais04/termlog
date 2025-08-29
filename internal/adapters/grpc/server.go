package grpc

import (
	"context"

	log "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/ports"
)

type Adapter struct {
	api ports.APIPort
	log.UnimplementedLogServiceServer
}

func NewAdapter(api ports.APIPort) *Adapter {
	return &Adapter{api: api}
}

func (a *Adapter) Log(ctx context.Context, req *log.LogEntry) (*log.LogResponse, error) {
	entry := domain.ReqToDomainLogEntry(req)
	err := a.api.Log(ctx, &entry)
	return &log.LogResponse{Success: err == nil}, err
}
