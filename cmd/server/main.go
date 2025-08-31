package server

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
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	gRPC "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	// memory "github.com/WillRabalais04/terminalLog/internal/adapters/memory" // prints outputs for testing purposes
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env found")
	}

	cache, err := database.NewRepo(&database.Config{
		Driver:       "sqlite3",
		DataSource:   "./data.db",
		SchemaString: "db/migrations/sqlite/000001_create_logs_table.up.sql",
	})
	if err != nil {
		log.Fatalf("couldn't access log cache: %v", err)
	}
	remote, err := database.NewRepo(&database.Config{
		Driver:       "pgx",
		DataSource:   utils.GetDSN(),
		SchemaString: "db/migrations/postgres/000001_create_logs_table.up.sql",
	})
	if err != nil {
		log.Fatalf("couldn't access remote log repo: %v", err)
	}

	repo := database.NewMultiRepo(cache, remote)
	svc := service.NewLogService(repo)
	grpcAdapter := gRPC.NewAdapter(svc)

	// run server
	listenPort := getEnvOrDefault("LISTEN_PORT", ":9090")
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

func getEnvOrDefault(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
