package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	gen "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/cmd/utils"
	db "github.com/WillRabalais04/terminalLog/internal/adapters/database"
	gRPC "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	// memory "github.com/WillRabalais04/terminalLog/internal/adapters/memory" // prints outputs for testing purposes
)

func main() {

	repo, err := db.GetRemoteRepo(utils.GetEnvOrDefault("DSN", utils.GetDSN("main")))
	if err != nil {
		log.Fatalf("server init failed: %v", err)
	}

	svc := service.NewLogService(repo)
	grpcAdapter := gRPC.NewServerAdapter(svc)

	// run server
	listenPort := utils.GetEnvOrDefault("LISTEN_PORT", ":9090")
	lis, err := net.Listen("tcp", listenPort)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", listenPort, err)
	}

	gRPCServer := grpc.NewServer()
	gen.RegisterLogServiceServer(gRPCServer, grpcAdapter)

	log.Println("grpc server listening on", listenPort)
	go func() {
		if err := gRPCServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve grpc server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down grpc server...")
	gRPCServer.GracefulStop()
}
