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
	prettyPrint("log", []*domain.LogEntry{entry}, err)
}

func testGet(ctx context.Context, svc service.LogService) {
	entry, err := svc.Get(ctx, "a68037cc-7293-4a03-a264-edc5ff05ba58")
	prettyPrint("get", []*domain.LogEntry{entry}, err)
}

func testList(ctx context.Context, svc service.LogService) {
	entries, err := svc.List(ctx, &ports.LogFilter{})
	prettyPrint("list", entries, err)
}

func testDelete(ctx context.Context, svc service.LogService) {
	err := svc.Delete(ctx, "a68037cc-7293-4a03-a264-edc5ff05ba58")
	prettyPrint("delete", nil, err)
}

func testDeleteMultiple(ctx context.Context, svc service.LogService) {
	filter := &ports.LogFilter{
		FilterTerms: map[string]ports.FilterValues{
			"git_repo": {Values: []string{"true"}},
		},
		FilterMode: ports.AND,
	}
	err := svc.DeleteMultiple(ctx, filter)
	if err != nil {
		log.Printf("testDeleteMultiple failed: %v", err)
	}
	prettyPrint("delete multiple", nil, err)
}

func prettyPrint(testName string, entries []*domain.LogEntry, err error) {
	fmt.Printf("---------------Test: %s---------------\n", testName)
	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		for _, e := range entries {
			domain.PrintLogEntry(e)
		}
	}
	spacing := make([]byte, len(testName))
	for range len(testName) + 6 {
		spacing = append(spacing, '-')
	}
	fmt.Printf("---------------%s---------------\n", spacing)
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

	testSVC := service.NewLogService(getLocalRepo())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testLog(ctx, *testSVC)
	testList(ctx, *testSVC)
	testGet(ctx, *testSVC)
	testDelete(ctx, *testSVC)
	testDeleteMultiple(ctx, *testSVC)
}
