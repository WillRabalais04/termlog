package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/WillRabalais04/terminalLog/cmd/utils"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	"github.com/WillRabalais04/terminalLog/internal/testutils"
)

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
		entries = append(entries, testutils.RandomLog())
	}

	if err := svc.Log(ctx, entries); err != nil {
		fmt.Printf("failed to log entry: %v\n", err)
	}
	testutils.PrettyPrint("log", entries, nil)
	return entries
}

func testGet(ctx context.Context, svc service.LogService) {
	filter := domain.NewFilterBuilder().
		AddSearchTerm("command", "pwd").
		Build()
	entries, err := svc.List(ctx, filter)
	if err != nil {
		log.Fatalf("couldn't get pwd dummy command: %v", err)
	}
	entry, err := svc.Get(ctx, entries[0].EventID)
	testutils.PrettyPrint("get", []*domain.LogEntry{entry}, err)
}

func testList(ctx context.Context, svc service.LogService) {
	filter1 := domain.NewFilterBuilder().
		AddFilterTerm("exit_code", "0").
		Build()
	entries1, err := svc.List(ctx, filter1)
	testutils.PrettyPrint("list 1", entries1, err)

	filter2 := domain.NewFilterBuilder().
		AddSearchTerm("command", "git status").
		Build()
	entries2, err := svc.List(ctx, filter2)
	testutils.PrettyPrint("list 2", entries2, err)
}

func testDelete(ctx context.Context, svc service.LogService) {
	filter := domain.NewFilterBuilder().
		AddSearchTerm("command", "pwd").
		Build()
	entries, err := svc.List(ctx, filter)
	if err != nil {
		log.Fatalf("couldn't get pwd dummy command: %v", err)
	}
	deleted, err := svc.Delete(ctx, entries[0].EventID)
	testutils.PrettyPrint("delete", []*domain.LogEntry{deleted}, err)
}

func testDeleteMultiple(ctx context.Context, svc service.LogService) {

	filter := domain.NewFilterBuilder().
		AddFilterTerm("git_repo", "true").
		Build()

	deleted, err := svc.DeleteMultiple(ctx, filter)
	if err != nil {
		log.Printf("testDeleteMultiple failed: %v", err)
	}
	testutils.PrettyPrint("delete multiple", deleted, err)
}

func main() {

	local, err := database.GetLocalRepo(utils.GetCachePath())
	if err != nil {
		log.Fatalf("couldn't init local repo: %v", err)
	}
	// remote, err := database.GetRemoteRepo(utils.GetTestDSN())
	// if err != nil {
	// 	log.Fatalf("couldn't init remote repo: %v", err)
	// }
	testSVC := service.NewLogService(local)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testLog(ctx, *testSVC, 6)
	testList(ctx, *testSVC)
	testGet(ctx, *testSVC)
	testDelete(ctx, *testSVC)
	testDeleteMultiple(ctx, *testSVC)
}
