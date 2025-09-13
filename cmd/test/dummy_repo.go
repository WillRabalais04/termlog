package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WillRabalais04/terminalLog/cmd/utils"
	"github.com/WillRabalais04/terminalLog/db"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/ports"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

/*
for reference:
var logColumns = []string{
	"event_id", "command", "exit_code", "ts", "shell_pid", "shell_uptime", "cwd",
	"prev_cwd", "user_name", "euid", "term", "hostname", "ssh_client",
	"tty", "git_repo", "git_repo_root", "git_branch", "git_commit",
	"git_status", "logged_successfully",
}
*/

func randomLog() *domain.LogEntry {
	return &domain.LogEntry{
		EventID:   uuid.New().String(),
		Command:   fmt.Sprintf("cmd-%d", rand.Intn(100)),
		ExitCode:  rand.Int31n(10),
		Timestamp: time.Now().UnixNano(),
		Shell_PID: int32(rand.Intn(5000)),
		// ShellUptime:          rand.Int63n(1000000),
		WorkingDirectory: fmt.Sprintf("/tmp/dir-%d", rand.Intn(5)),
		// PrevWorkingDirectory: "/" + uuid.New().String()[:8],
		User: "user" + uuid.New().String()[:8],
		EUID: int32(rand.Intn(1000)),
		// Term:                 uuid.New().String()[:8],
		// Hostname:             uuid.New().String()[:8],
		// SSHClient:            "",
		// TTY:                  fmt.Sprintf("pts/%d", rand.Intn(10)),
		GitRepo: rand.Intn(2) == 0,
		// GitRepoRoot:          "/" + uuid.New().String()[:8],
		// GitBranch:            uuid.New().String()[:8],
		// GitCommit:            uuid.New().String()[:8],
		// GitStatus:            "clean",
		// LoggedSuccessfully:   true,
	}
}

func testLog(ctx context.Context, svc service.LogService, n int) []*domain.LogEntry {
	var entries []*domain.LogEntry

	entries = append(entries, &domain.LogEntry{
		Command:            "pwd",
		ExitCode:           0,
		Timestamp:          time.Now().UnixNano(),
		WorkingDirectory:   "/home/test",
		User:               "tester",
		Hostname:           "knownhost",
		LoggedSuccessfully: true,
	})

	entries = append(entries, &domain.LogEntry{
		Command:            "ls",
		ExitCode:           1,
		Timestamp:          time.Now().UnixNano(),
		User:               "deleteme",
		Hostname:           "host",
		LoggedSuccessfully: true,
	})

	entries = append(entries, &domain.LogEntry{
		Command:            "git status",
		ExitCode:           0,
		Timestamp:          time.Now().UnixNano(),
		User:               "gituser",
		GitRepo:            true,
		GitRepoRoot:        "/repo",
		GitBranch:          "feature/test",
		GitCommit:          "abca",
		GitStatus:          "dirty",
		LoggedSuccessfully: true,
	})

	for range n - 3 {
		entries = append(entries, randomLog())
	}

	for _, entry := range entries {
		if err := svc.Log(ctx, entry); err != nil {
			fmt.Printf("failed to log entry %v: %v\n", entry.Command, err)
		}
	}

	prettyPrint("log", entries, nil)
	return entries
}

func testGet(ctx context.Context, svc service.LogService) { // can't make working test without getting logs first
	filter := ports.NewFilterBuilder().
		AddSearchTerm("command", "pwd").
		Build()
	entries, err := svc.List(ctx, filter)
	if err != nil {
		log.Fatalf("couldn't get pwd dummy command: %v", err)
	}
	entry, err := svc.Get(ctx, entries[0].EventID)
	prettyPrint("get", []*domain.LogEntry{entry}, err)
}

func testList(ctx context.Context, svc service.LogService) {
	filter1 := ports.NewFilterBuilder().
		AddFilterTerm("exit_code", "0").
		Build()
	entries1, err := svc.List(ctx, filter1)
	prettyPrint("list 1", entries1, err)

	filter2 := ports.NewFilterBuilder().
		AddSearchTerm("command", "git status").
		Build()
	entries2, err := svc.List(ctx, filter2)
	prettyPrint("list 2", entries2, err)
}

func testDelete(ctx context.Context, svc service.LogService) {
	filter := ports.NewFilterBuilder().
		AddSearchTerm("command", "pwd").
		Build()
	entries, err := svc.List(ctx, filter)
	if err != nil {
		log.Fatalf("couldn't get pwd dummy command: %v", err)
	}
	deleted, err := svc.Delete(ctx, entries[0].EventID)
	prettyPrint("delete", []*domain.LogEntry{deleted}, err)
}

func testDeleteMultiple(ctx context.Context, svc service.LogService) {

	filter := ports.NewFilterBuilder().
		AddFilterTerm("git_repo", "true").
		Build()

	deleted, err := svc.DeleteMultiple(ctx, filter)
	if err != nil {
		log.Printf("testDeleteMultiple failed: %v", err)
	}
	prettyPrint("delete multiple", deleted, err)
}

func prettyPrint(testName string, entries []*domain.LogEntry, err error) {
	width := 80
	label := fmt.Sprintf(" Test: %s ", testName)

	left := (width - len(label)) / 2
	right := width - len(label) - left

	fmt.Printf("%s%s%s\n", strings.Repeat("-", left), label, strings.Repeat("-", right))

	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		for _, e := range entries {
			domain.PrintLogEntry(e)
		}
	}

	fmt.Println(strings.Repeat("-", width))
}

func loadEnv() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}
	if err := godotenv.Load(filepath.Join(homeDir, ".termlogger", ".env")); err != nil {
		log.Println("no .env found, using system env vars")
	}
}
func getLocalRepo() *database.LogRepo {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("could not get home dir: %v", err)
	}
	cachePath := filepath.Join(utils.GetProjectRoot(homeDir), "cmd", "test", "logs", "cache.db")
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

func getRemoteRepo() *database.LogRepo {
	remote, err := database.NewRepo(&database.Config{
		Driver:       "pgx",
		DataSource:   utils.GetTestDSN(),
		SchemaString: db.PostgresSchema,
	})
	if err != nil {
		log.Fatalf("could not init remote repo (postgres): %v", err)
	}
	return remote
}
func getMultiRepo() *database.MultiRepo {
	multiRepo := database.NewMultiRepo(getLocalRepo(), getRemoteRepo())
	return multiRepo
}

func main() {
	loadEnv()

	testSVC := service.NewLogService(getRemoteRepo())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testLog(ctx, *testSVC, 6)
	testList(ctx, *testSVC)
	testGet(ctx, *testSVC)
	testDelete(ctx, *testSVC)
	testDeleteMultiple(ctx, *testSVC)
}
