package main

import (
	"encoding/binary"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

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

	entry := &pb.LogEntry{
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

	logFile := filepath.Join(homeDir, ".termlogger", "logs", "logs.pb")

	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	// testWithJSON(entry, getProjectRoot(homeDir)) // test by logging json files in this directory

	// marshalled proto files to feed to kafka later
	data, err := proto.Marshal(entry)

	if err != nil {
		log.Fatalf("failed to marshal: %v", err)
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	size := uint32(len(data))
	if err := binary.Write(file, binary.LittleEndian, size); err != nil {
		log.Fatalf("failed to write length: %v", err)
	}
	if _, err := file.Write(data); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}

}

func testWithJSON(newEntry *pb.LogEntry, projectDir string) {

	testLogsDir := filepath.Join(projectDir, "testing", "logs")
	testLogFile := filepath.Join(testLogsDir, "logs.jsonl")

	if err := os.MkdirAll(testLogsDir, 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	err := appendToJSONLines(newEntry, testLogFile)

	if err != nil {
		log.Fatalf("failed to write updated log file: %v", err)
	}
}
func appendToJSONLines(entry *pb.LogEntry, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	m := protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	jsonData, err := m.Marshal(entry)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(append(jsonData, '\n')); err != nil {
		return err
	}

	return nil
}

// utils

func getProjectRoot(homeDir string) string {
	configPath := filepath.Join(homeDir, ".termlogger", "project_root")
	projectRootBytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("failed to read project root config: %v", err)
	}
	return strings.TrimSpace(string(projectRootBytes))
}
