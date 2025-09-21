package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/joho/godotenv"
	"google.golang.org/protobuf/encoding/protojson"
)

func LoadEnv() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}
	if err := godotenv.Load(filepath.Join(homeDir, ".termlogger", ".env")); err != nil {
		log.Println("no .env found, using system env vars")
	}
}

func GetAppCachePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("error: could not get home dir: %v", err)
		return "cache.db" // keep logs in current directory worst case scenario
	}
	defaultPath := filepath.Join(homeDir, ".termlogger", "cache.db")
	return GetEnvOrDefault("CACHE_PATH", defaultPath)
}

func GetDSN(db string) string {
	var host, port, user, password, dbname, sslmode string
	switch db {
	case "main":
		host = GetEnvOrDefault("DB_HOST", "db")
		port = GetEnvOrDefault("DB_PORT", "5432")
		user = GetEnvOrDefault("DB_USER", "postgres")
		password = GetEnvOrDefault("DB_PASSWORD", "password")
		dbname = GetEnvOrDefault("DB_NAME", "logs")
		sslmode = GetEnvOrDefault("DB_SSLMODE", "disable")
	case "test":
		host = GetEnvOrDefault("TEST_DB_HOST", "localhost")
		port = GetEnvOrDefault("TEST_DB_PORT", "5434")
		user = GetEnvOrDefault("TEST_DB_USER", "test")
		password = GetEnvOrDefault("TEST_DB_PASSWORD", "test")
		dbname = GetEnvOrDefault("TEST_DB_NAME", "test_logs")
		sslmode = GetEnvOrDefault("TEST_DB_SSLMODE", "disable")
	case "unit_test":
		host = GetEnvOrDefault("UNIT_TEST_DB_HOST", "localhost")
		port = GetEnvOrDefault("UNIT_TEST_DB_PORT", "5435")
		user = GetEnvOrDefault("UNIT_TEST_DB_USER", "test")
		password = GetEnvOrDefault("UNIT_TEST_DB_PASSWORD", "test")
		dbname = GetEnvOrDefault("UNIT_TEST_DB_NAME", "unit_test_logs")
		sslmode = GetEnvOrDefault("UNIT_TEST_DB_SSLMODE", "disable")
	default:
		log.Fatal("invalid database provided for DSN string generation")
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func GetCachePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("could not get home dir: %v", err)
	}
	return filepath.Join(GetProjectRoot(homeDir), "cmd", "test", "logs", "cache.db")
}

func GetEnvOrDefault(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func LogJSON(newEntry *pb.LogEntry, projectDir string) {

	testLogFile := filepath.Join(projectDir, "test", "logs", "logs.json")

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

	for _, entry := range entries {
		jsonData, err := m.Marshal(entry)
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
