package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	gen "github.com/WillRabalais04/terminalLog/api/gen"
	grpc_adapter "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"
	print_adapter "github.com/WillRabalais04/terminalLog/internal/adapters/print"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
)

func main() {
	repoAdapter := print_adapter.NewAdapter()
	coreService := service.New(repoAdapter)
	grpcAdapter := grpc_adapter.NewAdapter(coreService)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen on port 9090: %v", err)
	}
	gRPCServer := grpc.NewServer()

	gen.RegisterLogServiceServer(gRPCServer, grpcAdapter)

	log.Println("gRPC server listening on port 9090...")

	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC server over port 9090: %v", err)
	}
}
