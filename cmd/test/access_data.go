package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/WillRabalais04/terminalLog/cmd/utils"
	"github.com/WillRabalais04/terminalLog/db"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/ports"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	"github.com/joho/godotenv"
)

func testLog(ctx context.Context, svc service.LogService) {
	entry := &domain.LogEntry{
		Command:              "aaa",
		ExitCode:             0,
		Timestamp:            458198549889,
		Shell_PID:            458198549,
		ShellUptime:          458198549889,
		WorkingDirectory:     "user/main",
		PrevWorkingDirectory: "user",
		User:                 "aaa",
		EUID:                 234,
		Term:                 "moew",
		Hostname:             "n/a",
		SSHClient:            "aaa",
		TTY:                  "asdf",
		IsGitRepo:            false,
		GitRepoRoot:          "0000",
		GitBranch:            "gfeat",
		GitCommit:            "sga",
		GitStatus:            "aa",
		LoggedSuccessfully:   true,
	}
	err := svc.Log(ctx, entry)

	prettyPrint("log", nil, err)
}

func testGet(ctx context.Context, svc service.LogService) {
	entry, err := svc.Get(ctx, 0)

	prettyPrint("get", []*domain.LogEntry{entry}, err)
}

func testList(ctx context.Context, svc service.LogService) {
	eventID := "a"
	command := "docker ps -a"
	entries, err := svc.List(ctx, &ports.LogQuery{
		EventID: &eventID,
		Command: &command,
		// User                 *string
		// ExitCode             *int
		// Limit                uint64
		// Offset               uint64
		// OrderBy              *string
		// GitRepo              *bool
		// Timestamp            *string
		// ShellPID             *uint64
		// ShellUptime          *uint64
		// WorkingDirectory     *string
		// PrevWorkingDirectory *string
		// EUID                 *uint64 // make uint32?
		// Term                 *string
		// Hostname             *string
		// SSHClient            *string
		// TTY                  *string
		// GitRepoRoot          *string
		// Branch               *string
		// Commit               *string
		// LoggedSuccesfully    *bool
	})

	if err != nil {
		log.Printf("testGet failed: %v", err)
	}
	fmt.Println(entries)
}

func testDelete(ctx context.Context, svc service.LogService) {
	err := svc.Delete(ctx, "aa")
	prettyPrint("delete", nil, err)
}

func testDeleteMultiple(ctx context.Context, svc service.LogService) {
	gitRepo := false
	err := svc.DeleteMultiple(ctx, &ports.LogQuery{
		GitRepo: &gitRepo,
	})
	if err != nil {
		log.Printf("testDeleteMultiple failed: %v", err)
	}
	prettyPrint("delete multiple", nil, err)

}

func prettyPrint(testName string, entries []*domain.LogEntry, err error) {
	log.Printf("---------------Test: %s---------------\n", testName)
	for _, e := range entries {
		domain.PrintLogEntry(e) // doesn't work
	}
	spacing := make([]byte, len(testName))
	for range len(testName) + 6 {
		spacing = append(spacing, '-')
	}
	log.Printf("---------------%s---------------\n", spacing)

}

func main() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}
	if err := godotenv.Load(filepath.Join(homeDir, ".termlogger", ".env")); err != nil {
		log.Println("no .env found, using system env vars")
	}
	var testSVC *service.LogService

	cachePath := filepath.Join(utils.GetProjectRoot(homeDir), "testing", "cache.db")
	cache, err := database.NewRepo(&database.Config{
		Driver:       "sqlite3",
		DataSource:   cachePath,
		SchemaString: db.SqliteSchema,
	})
	if err != nil {
		log.Printf("could not init cache repo (sqlite): %v", err)
		os.Exit(1)
	}

	// remote, err := database.NewRepo(&database.Config{
	// 	Driver:       "pgx",
	// 	DataSource:   utils.GetDSN(),
	// 	SchemaString: db.PostgresSchema,
	// })
	// if err != nil {
	// 	log.Fatalf("could not init remote repo (postgres): %v", err)
	// }
	// multiRepo := database.NewMultiRepo(cache, remote)

	// testSVC = service.NewLogService(multiRepo)
	testSVC = service.NewLogService(cache)
	// testSVC = service.NewLogService(remote)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testLog(ctx, *testSVC)
	testList(ctx, *testSVC)
	testGet(ctx, *testSVC)
	testDelete(ctx, *testSVC)
	testDeleteMultiple(ctx, *testSVC)

}
