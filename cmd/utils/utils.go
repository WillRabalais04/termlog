package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/db"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"google.golang.org/protobuf/encoding/protojson"
)

func GetLocalRepo(test bool) *database.LogRepo {
	var cachePath string
	if test {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("could not get home dir: %v", err)
		}
		cachePath = filepath.Join(GetProjectRoot(homeDir), "cmd", "test", "logs", "cache.db")
	}

	cache, err := database.NewRepo(&database.Config{
		Driver:       "sqlite3",
		DataSource:   cachePath,
		SchemaString: db.SqliteSchema,
	})
	if err != nil {
		log.Printf("could not init cache repo (sqlite): %v", err)
		os.Exit(1)
	}
	return cache
}

func GetRemoteRepo(test bool) *database.LogRepo {
	var dataSource string
	if test {
		dataSource = GetTestDSN()
	} else {
		dataSource = GetDSN()
	}
	remote, err := database.NewRepo(&database.Config{
		Driver:       "pgx",
		DataSource:   dataSource,
		SchemaString: db.PostgresSchema,
	})
	if err != nil {
		log.Fatalf("could not init remote repo (postgres): %v", err)
	}
	return remote
}
func GetMultiRepo(test bool) *database.MultiRepo {
	return database.NewMultiRepo(GetLocalRepo(test), GetRemoteRepo(test))
}

// func LoadEnv() {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		log.Fatalf("failed to get home directory: %v", err)
// 	}
// 	if err := godotenv.Load(filepath.Join(homeDir, ".termlogger", ".env")); err != nil {
// 		log.Println("no .env found, using system env vars")
// 	}
// }

func GetDSN() string {
	host := GetEnvOrDefault("DB_HOST", "localhost")
	port := GetEnvOrDefault("DB_PORT", "5432")
	user := GetEnvOrDefault("DB_USER", "postgres")
	password := GetEnvOrDefault("DB_PASSWORD", "password")
	dbname := GetEnvOrDefault("DB_NAME", "terminallog")
	sslmode := GetEnvOrDefault("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	return dsn
}

func GetTestDSN() string {
	host := GetEnvOrDefault("TEST_DB_HOST", "localhost")
	port := GetEnvOrDefault("TEST_DB_PORT", "5434")
	user := GetEnvOrDefault("TEST_DB_USER", "test")
	password := GetEnvOrDefault("TEST_DB_PASSWORD", "test")
	dbname := GetEnvOrDefault("TEST_DB_NAME", "test_logs")
	sslmode := GetEnvOrDefault("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	return dsn
}

func GetEnvOrDefault(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func LogJSON(newEntry *pb.LogEntry, projectDir string) {

	testLogFile := filepath.Join(projectDir, "cmd", "test", "logs", "logs.json")

	err := WriteToJSON(newEntry, testLogFile)
	if err != nil {
		log.Fatalf("failed to write updated log file: %v", err)
	}
}

func WriteToJSON(entry *pb.LogEntry, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var entries []*pb.LogEntry
	if len(fileBytes) > 0 {
		var rawEntries []json.RawMessage
		if err := json.Unmarshal(fileBytes, &rawEntries); err != nil {
			return err
		}

		for _, rawEntry := range rawEntries {
			var logEntry pb.LogEntry
			if err := protojson.Unmarshal(rawEntry, &logEntry); err != nil {
				return err
			}
			entries = append(entries, &logEntry)
		}
	}

	entries = append(entries, entry)

	var rawEntriesToMarshal []json.RawMessage
	m := protojson.MarshalOptions{
		EmitUnpopulated: true,
	}

	for _, e := range entries {
		jsonData, err := m.Marshal(e)
		if err != nil {
			return err
		}
		rawEntriesToMarshal = append(rawEntriesToMarshal, jsonData)
	}

	finalJSONData, err := json.MarshalIndent(rawEntriesToMarshal, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, finalJSONData, 0644)
}

func GetProjectRoot(homeDir string) string {
	configPath := filepath.Join(homeDir, ".termlogger", "project_root")
	projectRootBytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("failed to read project root config: %v", err)
	}
	return strings.TrimSpace(string(projectRootBytes))
}

func PrintLogEntry(entry *domain.LogEntry) {

	fmt.Printf("LogEntry: {\n")
	fmt.Printf("EventID:		%s\n", entry.EventID)
	fmt.Printf("Command:		%s\n", entry.Command)
	fmt.Printf("ExitCode:		%d\n", entry.ExitCode)
	fmt.Printf("Timestamp:		%d\n", entry.Timestamp)
	fmt.Printf("Shell_PID:		%d\n", entry.Shell_PID)
	fmt.Printf("ShellUptime:	%d\n", entry.ShellUptime)
	fmt.Printf("WorkingDirectory:		%s\n", entry.WorkingDirectory)
	fmt.Printf("PrevWorkingDirectory:		%s\n", entry.PrevWorkingDirectory)
	fmt.Printf("User:		%s\n", entry.User)
	fmt.Printf("EUID:		%d\n", entry.EUID)
	fmt.Printf("Term:		%s\n", entry.Term)
	fmt.Printf("Hostname:		%s\n", entry.Hostname)
	fmt.Printf("TTY:		%s\n", entry.TTY)
	fmt.Printf("GitRepo:		%t\n", entry.GitRepo)
	fmt.Printf("GitRepoRoot:		%s\n", entry.GitRepoRoot)
	fmt.Printf("GitBranch:		%s\n", entry.GitBranch)
	fmt.Printf("GitCommit:		%s\n", entry.GitCommit)
	fmt.Printf("GitStatus:		%s\n", entry.GitStatus)
	fmt.Printf("LoggedSuccessfully:		%t\n", entry.LoggedSuccessfully)
	fmt.Println("}")
}
