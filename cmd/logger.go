package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/WillRabalais04/terminalLog/cmd/utils"
	"github.com/WillRabalais04/terminalLog/db"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/service"
	"github.com/joho/godotenv"
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

	entry := &domain.LogEntry{
		Command:              *cmd,
		ExitCode:             int32(*exit),
		Timestamp:           *ts,
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
		GitRepo:            *isRepo,
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
	if err := godotenv.Load(filepath.Join(homeDir, ".termlogger", ".env")); err != nil {
		log.Println("no .env found, using system env vars")
	}

	cachePath := utils.GetEnvOrDefault("LOG_DIR", filepath.Join(homeDir, ".termlogger", "cache.db"))

	cache, err := database.NewRepo(&database.Config{
		Driver:       "sqlite3",
		DataSource:   cachePath,
		SchemaString: db.SqliteSchema,
	})
	if err != nil {
		log.Printf("could not init cache repo (sqlite): %v", err)
		os.Exit(1)
	}

	mode := os.Getenv("APP_MODE")

	var svc *service.LogService

	if mode == "org" {
		remote, err := database.NewRepo(&database.Config{
			Driver:       "pgx",
			DataSource:   utils.GetDSN(),
			SchemaString: db.PostgresSchema,
		})
		if err != nil {
			log.Fatalf("could not init remote repo (postgres): %v", err)
		}
		multiRepo := database.NewMultiRepo(cache, remote)
		svc = service.NewLogService(multiRepo)
	} else {
		svc = service.NewLogService(cache)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc.Log(ctx, entry)

	// test by logging json files in this directory
	// LogJSON(entry, getProjectRoot(homeDir))
}
