package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/WillRabalais04/terminalLog/internal/core/service"
	"github.com/joho/godotenv"

	gen "github.com/WillRabalais04/terminalLog/api/gen"
	gRPC "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"

	postgres "github.com/WillRabalais04/terminalLog/internal/adapters/postgres"
	sqlite "github.com/WillRabalais04/terminalLog/internal/adapters/sqlite"
	// memory "github.com/WillRabalais04/terminalLog/internal/adapters/memory" // prints outputs for testing purposes
)

const dbURL = "postgres://postgres:password@localhost:5433/logs?sslmode=disable"

func main() {
	if err := godotenv.Load(".env.local"); err != nil {
		log.Println("no .env.local found, using system env vars")
	}
	db, err := sqlite.InitDB(dbURL)
	mode := os.Getenv("MODE")
	if mode == "org" {
		db, err = postgres.InitDB(dbURL)
	}

	// init / connect to DB
	if err != nil {
		log.Fatalf("Failed to initialize db: %v", err)
	}

	// repo, err := print.NewAdapter() (db)
	repo, err := postgres.NewRepository(db)
	if err != nil {
		log.Fatalf("Failed to connect to repository: %v", err)
	}

	// init services & adapters
	coreService := service.NewLogService(repo)
	grpcAdapter := gRPC.NewAdapter(coreService)

	// run server
	listenPort := ":9090"
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
