package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"google.golang.org/protobuf/encoding/protojson"
)

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

func GetEnvOrDefault(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func LogJSON(newEntry *pb.LogEntry, projectDir string) {

	testLogsDir := filepath.Join(projectDir, "testing", "logs")
	testLogFile := filepath.Join(testLogsDir, "logs.json")

	if err := os.MkdirAll(testLogsDir, 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

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
