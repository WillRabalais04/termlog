package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/WillRabalais04/terminalLog/cmd/utils"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	grpcAdapter "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	"github.com/WillRabalais04/terminalLog/internal/testutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func testLog(ctx context.Context, svc service.LogService, n int) {
	var entries []*domain.LogEntry

	entries = append(entries, &domain.LogEntry{
		Command:          "pwd",
		ExitCode:         0,
		Timestamp:        time.Now().UnixNano(),
		WorkingDirectory: "/home/client",
		User:             "grpc_user",
		Hostname:         "client_host",
	})
	entries = append(entries, &domain.LogEntry{
		Command:   "ls",
		ExitCode:  1,
		Timestamp: time.Now().UnixNano(),
		User:      "deleteme_client",
		Hostname:  "host",
	})
	entries = append(entries, &domain.LogEntry{
		Command:     "git status",
		ExitCode:    69,
		Timestamp:   time.Now().UnixNano(),
		User:        "gituser_client",
		GitRepo:     true,
		GitRepoRoot: "/client/repo",
		GitBranch:   "main",
	})

	for i := 0; i < n-3; i++ {
		entries = append(entries, testutils.RandomLog())
	}

	if err := svc.Log(ctx, entries); err != nil {
		fmt.Printf("failed to log entries: %v\n", err)
	}

	testutils.PrettyPrint("log", entries, nil)
}

func testGet(ctx context.Context, svc service.LogService) {
	filter := domain.NewFilterBuilder().
		AddSearchTerm("command", "pwd").
		Build()
	entries, err := svc.List(ctx, filter)
	if err != nil || len(entries) == 0 {
		log.Fatalf("couldn't find 'pwd' entry to get: %v", err)
	}
	entry, err := svc.Get(ctx, entries[0].EventID)
	testutils.PrettyPrint("get", []*domain.LogEntry{entry}, err)
}

func testList(ctx context.Context, svc service.LogService) {
	filter1 := domain.NewFilterBuilder().
		AddFilterTerm("exit_code", "0").
		Build()
	entries1, err := svc.List(ctx, filter1)
	testutils.PrettyPrint("list (exit_code=0)", entries1, err)

	filter2 := domain.NewFilterBuilder().
		AddSearchTerm("user", "client").
		Build()
	entries2, err := svc.List(ctx, filter2)
	testutils.PrettyPrint("list (user contains 'client')", entries2, err)
}

func testDelete(ctx context.Context, svc service.LogService) {
	filter := domain.NewFilterBuilder().
		AddSearchTerm("command", "pwd").
		Build()
	entries, err := svc.List(ctx, filter)
	if err != nil || len(entries) == 0 {
		log.Fatalf("couldn't find 'pwd' entry to delete: %v", err)
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
	testutils.PrettyPrint("delete multiple (git_repo=true)", deleted, err)
}

func main() {
	port := utils.GetEnvOrDefault("API_HOST_PORT", "localhost:9090")
	conn, err := grpc.Dial(port, grpc.WithTransportCredentials(insecure.NewCredentials()))

	var repo *service.LogService

	if err == nil {
		log.Println("✅ Connected to grpc server successfully.")
		clientAdapter := grpcAdapter.NewClientAdapter(conn)
		repo = service.NewLogService(clientAdapter)
		defer conn.Close()
	} else {
		log.Println("could not connect to grpc server, falling back to local cache:", err)
		localRepo, localErr := database.GetLocalRepo(utils.GetCachePath())
		if localErr != nil {
			fmt.Println("failed to init local repo:", localErr)
			os.Exit(1)
		}
		repo = service.NewLogService(localRepo)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testLog(ctx, *repo, 6)
	testList(ctx, *repo)
	testGet(ctx, *repo)
	testDelete(ctx, *repo)
	testDeleteMultiple(ctx, *repo)
}
