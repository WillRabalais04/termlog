package test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/WillRabalais04/terminalLog/cmd/utils"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestMain(m *testing.M) {
	utils.LoadEnv()
	log.Println("Setting up unit tests database...")
	cmd := exec.Command("make", "migrate-up-unit-tests")
	cmd.Dir = "../../"
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("Failed to setup test database: %v\nOutput:\n%s", err, string(output))
	}

	waitForDB(utils.GetDSN("unit_test"))
	exitCode := m.Run()

	log.Println("Tearing down database...")
	cmd = exec.Command("make", "stop-db-unit-tests")
	cmd.Dir = "../../"
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Warning: failed to tear down test database: %v\nOutput:\n%s", err, string(output))
	}

	os.Exit(exitCode)
}

// runs standard tests for cache, remote and multi repos
func TestRepositoryImplementations(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T) (*service.LogService, func())
	}{
		{"LocalRepo", setupLocalTest},
		{"RemoteRepo", setupRemoteTest},
		{"MultiRepo", setupMultiRepoTest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name != "LocalRepo" { // clean db for remote repos
				db, err := sql.Open("pgx", utils.GetDSN("unit_test"))
				if err != nil {
					t.Fatalf("failed to connect for truncation: %v", err)
				}
				defer db.Close()
				_, err = db.Exec("TRUNCATE TABLE logs RESTART IDENTITY")
				if err != nil {
					t.Fatalf("failed to truncate logs table: %v", err)
				}
			}

			svc, teardown := tc.setup(t)
			defer teardown()

			t.Run("List from Empty Repository", func(t *testing.T) {
				results, err := svc.List(context.Background(), &domain.LogFilter{})
				if err != nil {
					t.Fatalf("List from empty repo failed: %v", err)
				}
				if len(results) != 0 {
					t.Errorf("Expected 0 entries from empty repo, but got %d", len(results))
				}
			})

			runStandardTests(t, svc)
		})
	}
}

func TestMultiRepoLocalBackupSync(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("Setting up multi-repo remote backup test...")
	runMakeCommand(t, "migrate-up-unit-tests")

	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "test_cache.db")

	multiRepo, err := database.GetMultiRepo(cachePath, utils.GetDSN("unit_test"))
	if err != nil {
		t.Fatalf("could not init multi repo: %v", err)
	}
	svc := service.NewLogService(multiRepo)

	onlineEntry := &domain.LogEntry{Command: "log-when-online", ExitCode: 0}
	if err := svc.Log(ctx, []*domain.LogEntry{onlineEntry}); err != nil {
		t.Fatalf("Failed to log entry while DB was online: %v", err)
	}
	t.Log("Stopping remote DB to test local backup...")
	runMakeCommand(t, "stop-db-unit-tests")

	offlineEntry1 := &domain.LogEntry{Command: "offline-entry-1", ExitCode: 0}
	offlineEntry2 := &domain.LogEntry{Command: "offline-entry-2", ExitCode: 1}
	if err := svc.Log(ctx, []*domain.LogEntry{offlineEntry1}); err != nil {
		t.Fatalf("Fallback failed for offline-entry-1: %v", err)
	}
	if err := svc.Log(ctx, []*domain.LogEntry{offlineEntry2}); err != nil {
		t.Fatalf("Fallback failed for offline-entry-2: %v", err)
	}
	t.Log("Successfully logged 2 entries to cache while remote was down.")
	t.Log("Restarting remote DB to test sync...")

	runMakeCommand(t, "migrate-up-unit-tests")

	triggerEntry := &domain.LogEntry{Command: "trigger-sync-log", ExitCode: 0}
	if err := svc.Log(ctx, []*domain.LogEntry{triggerEntry}); err != nil {
		t.Fatalf("Failed to log trigger entry: %v", err)
	}

	if _, err := multiRepo.FlushCache(ctx); err != nil { // manual cache flushing bc logger binary would do it asynchronously but just calling the log function does not
		t.Fatalf("Manual cache flush failed: %v", err)
	}
	t.Log("Logged trigger entry and flushed cache, sync should have occurred.")
	remoteSVC := service.NewLogService(multiRepo.GetRemote())

	allRemoteLogs, err := remoteSVC.List(ctx, &domain.LogFilter{})
	if err != nil {
		t.Fatalf("Failed to list logs from remote repo for verification: %v", err)
	}

	expectedCommands := map[string]bool{
		"offline-entry-1":  false,
		"offline-entry-2":  false,
		"trigger-sync-log": false,
	}

	for _, logEntry := range allRemoteLogs {
		if _, exists := expectedCommands[logEntry.Command]; exists {
			expectedCommands[logEntry.Command] = true
		}
	}
	for command, found := range expectedCommands {
		if !found {
			t.Errorf("Verification failed: Expected to find command '%s' in remote DB, but it was missing.", command)
		}
	}
	t.Log("Verification successful: All expected logs are present in the remote database.")

	localSVC := service.NewLogService(multiRepo.GetCache())
	cachedLogs, err := localSVC.List(ctx, &domain.LogFilter{})
	if err != nil {
		t.Fatalf("Failed to list logs from local cache for verification: %v", err)
	}
	if len(cachedLogs) != 0 {
		t.Fatalf("Verification failed: Expected local cache to be empty, but found %d logs.", len(cachedLogs))
	}
	t.Log("Verification successful: Local cache has been flushed.")

	svc.DeleteMultiple(ctx, &domain.LogFilter{}) // clear
	runMakeCommand(t, "stop-db-unit-tests")
}

// runStandardTests has tests that each repo type must pass
func runStandardTests(t *testing.T, svc *service.LogService) {
	ctx := context.Background()

	dummyEntries := []*domain.LogEntry{
		{Command: "ls -l", ExitCode: 0, Timestamp: time.Now().UnixNano()},
		{Command: "cat file.txt", ExitCode: 1, Timestamp: time.Now().UnixNano()},
		{Command: "git status", ExitCode: 0, GitRepo: true, Timestamp: time.Now().UnixNano()},
		{Command: "git push", ExitCode: 0, GitRepo: true, Timestamp: time.Now().UnixNano()},
	}
	for _, entry := range dummyEntries {
		if err := svc.Log(ctx, []*domain.LogEntry{entry}); err != nil {
			t.Fatalf("Failed to log seed entry for command '%s': %v", entry.Command, err)
		}
	}

	t.Run("Get Operations", func(t *testing.T) {
		t.Run("Get Existing Entry", func(t *testing.T) {
			entryToGet := &domain.LogEntry{Command: "get-me"}
			err := svc.Log(ctx, []*domain.LogEntry{entryToGet})
			if err != nil {
				t.Fatalf("Failed to log entry: %v", err)
			}

			foundEntry, err := svc.Get(ctx, entryToGet.EventID)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}
			if foundEntry.Command != "get-me" {
				t.Errorf("Expected command 'get-me', got '%s'", foundEntry.Command)
			}
		})

		t.Run("Get Non-Existent Entry", func(t *testing.T) {
			nonExistentID := "00000000-0000-0000-0000-000000000000"
			_, err := svc.Get(ctx, nonExistentID)
			if err == nil {
				t.Error("Expected an error when getting a non-existent entry, but got nil")
			}
		})
	})

	t.Run("List with Filters", func(t *testing.T) {
		t.Run("Filter by Search Term", func(t *testing.T) {
			filter := domain.NewFilterBuilder().AddSearchTerm("command", "git").Build()
			results, err := svc.List(ctx, filter)
			if err != nil {
				t.Fatalf("List with search term failed: %v", err)
			}
			if len(results) != 2 {
				t.Errorf("Expected 2 entries for command containing 'git', but got %d", len(results))
			}
		})

		t.Run("Filter by Exact Term", func(t *testing.T) {
			filter := domain.NewFilterBuilder().AddFilterTerm("exit_code", "0").Build()
			results, err := svc.List(ctx, filter)
			if err != nil {
				t.Fatalf("List with filter term failed: %v", err)
			}
			if len(results) != 4 {
				t.Errorf("Expected 4 entries with exit_code 0, but got %d", len(results))
			}
		})

		t.Run("Filter with No Matches", func(t *testing.T) {
			filter := domain.NewFilterBuilder().AddSearchTerm("command", "nonexistentcommand").Build()
			results, err := svc.List(ctx, filter)
			if err != nil {
				t.Fatalf("List with no matches failed: %v", err)
			}
			if len(results) != 0 {
				t.Errorf("Expected 0 entries, but got %d", len(results))
			}
		})

		t.Run("Filter with Multiple Combined Terms", func(t *testing.T) {
			filter := domain.NewFilterBuilder().
				AddFilterTerm("git_repo", "true").
				AddFilterTerm("exit_code", "0").
				Build()
			results, err := svc.List(ctx, filter)
			if err != nil {
				t.Fatalf("List with combined filters failed: %v", err)
			}
			if len(results) != 2 {
				t.Errorf("Expected 2 entries that are git_repo and exit_code 0, but got %d", len(results))
			}
		})
	})

	t.Run("Delete Operations", func(t *testing.T) {
		t.Run("Delete Single Entry", func(t *testing.T) {
			entryToDelete := &domain.LogEntry{Command: "delete-me"}
			err := svc.Log(ctx, []*domain.LogEntry{entryToDelete})
			if err != nil {
				t.Fatalf("Failed to log entry for deletion: %v", err)
			}

			_, err = svc.Delete(ctx, entryToDelete.EventID)
			if err != nil {
				t.Fatalf("Delete failed: %v", err)
			}

			_, err = svc.Get(ctx, entryToDelete.EventID)
			if err == nil {
				t.Error("Expected an error when getting a deleted entry, but it still exists.")
			}
		})

		t.Run("Delete Multiple Entries", func(t *testing.T) {
			filter := domain.NewFilterBuilder().AddFilterTerm("git_repo", "true").Build()
			deleted, err := svc.DeleteMultiple(ctx, filter)
			if err != nil {
				t.Fatalf("DeleteMultiple failed: %v", err)
			}
			if len(deleted) != 2 {
				t.Fatalf("Expected to delete 2 git repo entries, but deleted %d", len(deleted))
			}
			remaining, err := svc.List(ctx, filter)
			if err != nil {
				t.Fatalf("List after DeleteMultiple failed: %v", err)
			}
			if len(remaining) != 0 {
				t.Errorf("Expected 0 git repo entries after deletion, but found %d", len(remaining))
			}
		})
	})
}

func setupLocalTest(t *testing.T) (*service.LogService, func()) {
	t.Helper()

	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "test_cache.db")

	repo, err := database.GetLocalRepo(cachePath)
	if err != nil {
		t.Fatalf("could not init local repo: %v", err)
	}
	svc := service.NewLogService(repo)

	return svc, func() {} // no-op bc tempdir cleans dir
}

func setupRemoteTest(t *testing.T) (*service.LogService, func()) {
	t.Helper()

	repo, err := database.GetRemoteRepo(utils.GetDSN("unit_test"))
	if err != nil {
		t.Fatalf("could not init remote repo: %v", err)
	}
	svc := service.NewLogService(repo)

	return svc, func() {} // no-op bc teardown is handled in testmain
}

func setupMultiRepoTest(t *testing.T) (*service.LogService, func()) {
	t.Helper()

	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "test_cache.db")

	multiRepo, err := database.GetMultiRepo(cachePath, utils.GetDSN("unit_test"))
	if err != nil {
		t.Fatalf("Could not init multi repo: %v", err)
	}
	svc := service.NewLogService(multiRepo)

	return svc, func() {} // no-op bc teardown is handled in testmain
}

func waitForDB(dsn string) {
	var db *sql.DB
	var err error
	retries := 10
	for i := 0; i < retries; i++ {
		db, err = sql.Open("pgx", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Database connection successful.")
				db.Close()
				return
			}
		}
		log.Printf("... DB not ready, retrying in 1 second (%d/%d)", i+1, retries)
		time.Sleep(1 * time.Second)
	}
	log.Fatalf("Could not connect to the database after %d retries: %v", retries, err)
}

func runMakeCommand(t *testing.T, args ...string) {
	t.Helper()
	cmd := exec.Command("make", args...)
	cmd.Dir = "../../"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run 'make %s' from root: %v\nOutput:\n%s", args, err, string(output))
	}
}
