package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/joho/godotenv"

	gen "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/cmd/utils"
	gRPC "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	// memory "github.com/WillRabalais04/terminalLog/internal/adapters/memory" // prints outputs for testing purposes
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env found")
	}

	repo := utils.GetMultiRepo(false)
	svc := service.NewLogService(repo)
	grpcAdapter := gRPC.NewAdapter(svc)

	// run server
	listenPort := utils.GetEnvOrDefault("LISTEN_PORT", ":9090")
	lis, err := net.Listen("tcp", listenPort)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", listenPort, err)
	}

	gRPCServer := grpc.NewServer()
	gen.RegisterLogServiceServer(gRPCServer, grpcAdapter)

	log.Println("gRPC server listening on", listenPort)
	go func() {
		if err := gRPCServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gRPC server...")
	gRPCServer.GracefulStop()
}
