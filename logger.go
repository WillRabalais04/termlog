package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	pg "github.com/WillRabalais04/terminalLog/internal/adapters/postgres"
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

	db, err := pg.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize db: %v", err)
	}

	// repo, err := print.NewAdapter() (db)
	repo, err := pg.NewRepository(db) // doesn't pass by api call yet but might in future
	if err != nil {
		log.Fatalf("Failed to connect to repository: %v\n Logging locally...", err)
		// repo, err = print.NewAdapter()
		if err != nil {
			log.Fatalf("Could not cache logs locally (%v). Log lost...", err)
		}
	}

	coreService := service.NewLogService(repo)
	coreService.Log(entry)
	// init services & adapters

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
