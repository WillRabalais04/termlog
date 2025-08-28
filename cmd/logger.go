package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/service"

	"google.golang.org/protobuf/encoding/protojson"
)

const dbURL = "postgres://postgres:password@localhost:5433/logs?sslmode=disable"

func main() {

	cmd := flag.String("cmd", "", "Command executed")
	exit := flag.Int("exit", 0, "Exit code of command")
	ts := flag.Int64("ts", time.Now().Unix(), "Unix timestamp")
	spid := flag.Int("pid", 0, "Shell PID")
	uptime := flag.Int64("uptime", 0, "Shell uptime in seconds")
	cwd := flag.String("cwd", "", "Current working directory")
	oldpwd := flag.String("oldpwd", "", "Previous working directory")
	user := flag.String("user", "", "Username")
	euid := flag.Int("euid", 0, "Effective UID")
	term := flag.String("term", "", "Terminal type")
	hostname := flag.String("hostname", "", "Hostname")
	sshClient := flag.String("ssh", "", "SSH client info")
	tty := flag.String("tty", "", "TTY")
	isRepo := flag.Bool("gitrepo", false, "Is inside git repo")
	gitRoot := flag.String("gitroot", "", "Git repo root")
	gitBranch := flag.String("gitbranch", "", "Git branch")
	gitCommit := flag.String("gitcommit", "", "Git commit hash")
	gitStatus := flag.String("gitstatus", "", "Git status")

	flag.Parse()

	// entry := &pb.LogEntry{
	// 	Command:              *cmd,
	// 	ExitCode:             int32(*exit),
	// 	Timestamp:            *ts,
	// 	Shell_PID:            int32(*spid),
	// 	ShellUptime:          *uptime,
	// 	WorkingDirectory:     *cwd,
	// 	PrevWorkingDirectory: *oldpwd,
	// 	User:                 *user,
	// 	EUID:                 int32(*euid),
	// 	Term:                 *term,
	// 	Hostname:             *hostname,
	// 	SSHClient:            *sshClient,
	// 	TTY:                  *tty,
	// 	IsGitRepo:            *isRepo,
	// 	GitRepoRoot:          *gitRoot,
	// 	GitBranch:            *gitBranch,
	// 	GitCommit:            *gitCommit,
	// 	GitStatus:            *gitStatus,
	// 	LoggedSuccessfully:   true,
	// }

	entry := &domain.LogEntry{
		Command:              *cmd,
		ExitCode:             int32(*exit),
		Timestamp:            *ts,
		Shell_PID:            int32(*spid),
		ShellUptime:          *uptime,
		WorkingDirectory:     *cwd,
		PrevWorkingDirectory: *oldpwd,
		User:                 *user,
		EUID:                 int32(*euid),
		Term:                 *term,
		Hostname:             *hostname,
		SSHClient:            *sshClient,
		TTY:                  *tty,
		IsGitRepo:            *isRepo,
		GitRepoRoot:          *gitRoot,
		GitBranch:            *gitBranch,
		GitCommit:            *gitCommit,
		GitStatus:            *gitStatus,
		LoggedSuccessfully:   true,
	}

	if !entry.LoggedSuccessfully {
		log.Fatal("unsuccessfully logged")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}

	// db, err := database.InitDB(&database.Config{
	// 	Driver:     "sqlite3",
	// 	DataSource: "./data.db",
	// 	SchemaFile: "db/sqlite/000001_create_logs_table.up.sql",
	// })
	// if err != nil {
	// 	log.Fatalf("Failed to initialize db: %v", err)
	// }

	// repo, err := print.NewAdapter() (db)
	// repo, err := database.NewRepo(&database.Config{
	// 	Driver:     "sqlite3",
	// 	DataSource: "./data.db",
	// 	SchemaFile: "db/sqlite/000001_create_logs_table.up.sql",
	// })
	// coreService := service.NewLogService(repo)
	// coreService.Log(entry)
	// init services & adapters

	cacheRepo, err := database.NewRepo(&database.Config{ // add error handlng and clean up a lot
		Driver:     "sqlite3",
		DataSource: "./data.db",
		SchemaFile: "db/sqlite/000001_create_logs_table.up.sql",
	})

	var svc *service.LogService

	isOrgMode, err := strconv.ParseBool(os.Getenv("ORG_MODE"))
	if isOrgMode {
		remoteRepo, err := database.NewRepo(&database.Config{
			Driver:     "pgx",
			DataSource: getDSN(),
			SchemaFile: "db/migrations/postgres/000001_create_logs_table.up.sql",
		})
		if err != nil {
			log.Fatalf("couldn't access remote log repo: %v", err)
		}
		multiRepo := database.NewMultiRepo(cacheRepo, remoteRepo)
		svc = service.NewLogService(multiRepo)

	} else {
		svc = service.NewLogService(cacheRepo)

	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second) // figure this out

	svc.Log(ctx, entry)

	logFile := filepath.Join(homeDir, ".termlogger", "logs", "logs.pb")

	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	// test by logging json files in this directory
	// logJSON(entry, getProjectRoot(homeDir))

	// marshalled proto files to feed to kafka later
	// data, err := proto.Marshal(entry)

	// if err != nil {
	// 	log.Fatalf("failed to marshal: %v", err)
	// }

	// file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	// if err != nil {
	// 	log.Fatalf("failed to open file: %v", err)
	// }
	// defer file.Close()

	// size := uint32(len(data))
	// if err := binary.Write(file, binary.LittleEndian, size); err != nil {
	// 	log.Fatalf("failed to write length: %v", err)
	// }
	// if _, err := file.Write(data); err != nil {
	// 	log.Fatalf("failed to write data: %v", err)
	// }

}

func logJSON(newEntry *pb.LogEntry, projectDir string) {

	testLogsDir := filepath.Join(projectDir, "testing", "logs")
	testLogFile := filepath.Join(testLogsDir, "logs.json")

	if err := os.MkdirAll(testLogsDir, 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	err := writeToJSON(newEntry, testLogFile)
	if err != nil {
		log.Fatalf("failed to write updated log file: %v", err)
	}
}

// utils
func writeToJSON(entry *pb.LogEntry, path string) error {
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

func getProjectRoot(homeDir string) string {
	configPath := filepath.Join(homeDir, ".termlogger", "project_root")
	projectRootBytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("failed to read project root config: %v", err)
	}
	return strings.TrimSpace(string(projectRootBytes))
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
