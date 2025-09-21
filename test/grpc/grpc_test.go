package grpc_test

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	grpcClient "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"
	grpcServer "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	"github.com/WillRabalais04/terminalLog/internal/testutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var testSvc *service.LogService

func TestMain(m *testing.M) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()

	tempDir, err := os.MkdirTemp("", "client-test-cache")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	cachePath := filepath.Join(tempDir, "test_cache.db")
	localRepo, err := database.GetLocalRepo(cachePath)
	if err != nil {
		log.Fatalf("Failed to init local repo for test server: %v", err)
	}

	serverSvc := service.NewLogService(localRepo)
	pb.RegisterLogServiceServer(s, grpcServer.NewServerAdapter(serverSvc))

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	clientAdapter := grpcClient.NewClientAdapter(conn)
	testSvc = service.NewLogService(clientAdapter)

	exitCode := m.Run()
	s.GracefulStop()
	os.Exit(exitCode)
}

func TestClientServerInteraction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := testSvc.DeleteMultiple(ctx, &domain.LogFilter{}); err != nil {
		t.Fatalf("Failed to clean database before test: %v", err)
	}

	var loggedEntries []*domain.LogEntry

	t.Run("Log Entries", func(t *testing.T) {
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
			Command:   "git status",
			ExitCode:  0,
			Timestamp: time.Now().UnixNano(),
			User:      "gituser_client",
			GitRepo:   true,
		})
		for i := 0; i < 4; i++ {
			entries = append(entries, testutils.RandomLog())
		}

		if err := testSvc.Log(ctx, entries); err != nil {
			t.Fatalf("Log request failed: %v", err)
		}

		allEntries, err := testSvc.List(ctx, &domain.LogFilter{})
		if err != nil {
			t.Fatalf("Failed to list entries after logging: %v", err)
		}
		if len(allEntries) != len(entries) {
			t.Errorf("Expected %d logged entries, but found %d", len(entries), len(allEntries))
		}
		loggedEntries = allEntries
	})

	t.Run("Get Existing Entry", func(t *testing.T) {
		if len(loggedEntries) == 0 {
			t.Skip("Skipping test, no entries were logged in the previous step.")
		}
		entryToGetID := loggedEntries[0].EventID
		entry, err := testSvc.Get(ctx, entryToGetID)
		if err != nil {
			t.Fatalf("Get request failed for existing ID %s: %v", entryToGetID, err)
		}
		if entry.EventID != entryToGetID {
			t.Errorf("Get request returned wrong entry. Expected ID %s, got %s", entryToGetID, entry.EventID)
		}
	})

	t.Run("List with Filters", func(t *testing.T) {
		filter := domain.NewFilterBuilder().AddFilterTerm("exit_code", "0").Build()
		results, err := testSvc.List(ctx, filter)
		if err != nil {
			t.Fatalf("List with filter failed: %v", err)
		}

		expectedCount := 0
		for _, entry := range loggedEntries {
			if entry.ExitCode == 0 {
				expectedCount++
			}
		}
		if len(results) != expectedCount {
			t.Errorf("Expected %d entries with exit_code 0, but got %d", expectedCount, len(results))
		}
	})

	t.Run("Delete Entry", func(t *testing.T) {
		if len(loggedEntries) < 2 {
			t.Skip("Skipping test, not enough entries were logged.")
		}
		entryToDeleteID := loggedEntries[1].EventID
		deleted, err := testSvc.Delete(ctx, entryToDeleteID)
		if err != nil {
			t.Fatalf("Delete request failed: %v", err)
		}
		if deleted.EventID != entryToDeleteID {
			t.Errorf("Delete returned wrong entry. Expected ID %s, got %s", entryToDeleteID, deleted.EventID)
		}

		_, err = testSvc.Get(ctx, entryToDeleteID)
		if err == nil {
			t.Errorf("Expected an error when getting a deleted entry, but it still exists.")
		}
	})

	t.Run("Delete Multiple Entries", func(t *testing.T) {
		gitEntries := []*domain.LogEntry{
			{Command: "git commit", GitRepo: true},
			{Command: "git push", GitRepo: true},
		}
		if err := testSvc.Log(ctx, gitEntries); err != nil {
			t.Fatalf("Failed to log extra git entries for deletion test: %v", err)
		}

		filter := domain.NewFilterBuilder().AddFilterTerm("git_repo", "true").Build()

		entriesBefore, err := testSvc.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list git entries before deletion: %v", err)
		}
		if len(entriesBefore) == 0 {
			t.Fatal("Expected to find git entries to delete, but found none.")
		}

		deleted, err := testSvc.DeleteMultiple(ctx, filter)
		if err != nil {
			t.Fatalf("DeleteMultiple request failed: %v", err)
		}
		if len(deleted) != len(entriesBefore) {
			t.Errorf("Expected to delete %d entries, but DeleteMultiple returned %d", len(entriesBefore), len(deleted))
		}

		entriesAfter, err := testSvc.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list git entries after deletion: %v", err)
		}
		if len(entriesAfter) != 0 {
			t.Errorf("Expected 0 git entries after deletion, but found %d", len(entriesAfter))
		}
	})
}
