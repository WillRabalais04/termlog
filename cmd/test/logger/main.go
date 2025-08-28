package logger_test

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"

	"github.com/WillRabalais04/terminalLog/internal/core/service"
	"github.com/joho/godotenv"

	gen "github.com/WillRabalais04/terminalLog/api/gen"

	// postgres "github.com/WillRabalais04/terminalLog/internal/adapters/postgres"
	database "github.com/WillRabalais04/terminalLog/internal/adapters/db"
	// memory "github.com/WillRabalais04/terminalLog/internal/adapters/memory" // prints outputs for testing purposes
)

func main() {

	if err := godotenv.Load("../.env.local"); err != nil {
		log.Println("no .env.local found, using system env vars")
	}

	cache, err := database.NewRepo(&database.Config{
		Driver:     "sqlite3",
		DataSource: "./data.db",
		SchemaFile: "db/migrations/sqlite/000001_create_logs_table.up.sql",
	})
	if err != nil {
		log.Fatalf("couldn't access log cache: %v", err)
	}

	isOrgMode, err := strconv.ParseBool(os.Getenv("ORG_MODE"))
	if isOrgMode {
		remoteDB, err := database.NewRepo(&database.Config{
			Driver:     "pgx",
			DataSource: getDSN(),
			SchemaFile: "db/migrations/postgres/000001_create_logs_table.up.sql",
		})
		if err != nil {
			log.Fatalf("couldn't access remote log repo: %v", err)
		}
	}

	
	logService := service.NewLogService(nil, mainDB)

	// init services & adapters
	// cacheService := service.NewLogService(cache)
	// cacheService := service.NewLogService(mainDBService)
	// grpcAdapter := gRPC.NewAdapter(coreService)

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

func getDSN() string {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "")
	dbname := getEnvOrDefault("DB_NAME", "terminallog")
	sslmode := getEnvOrDefault("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	return dsn
}

func getEnvOrDefault(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
