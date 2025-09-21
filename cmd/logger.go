package main

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/WillRabalais04/terminalLog/cmd/utils"
	"github.com/WillRabalais04/terminalLog/db"
	"github.com/WillRabalais04/terminalLog/internal/adapters/database"
	grpcAdapter "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	utils.LoadEnv()

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
	jsonMode := flag.Bool("json", false, "Log output to a JSON file in addition to the database.")

	flag.Parse()

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
		GitRepo:              *isRepo,
		GitRepoRoot:          *gitRoot,
		GitBranch:            *gitBranch,
		GitCommit:            *gitCommit,
		GitStatus:            *gitStatus,
		LoggedSuccessfully:   true,
	}

	if !entry.LoggedSuccessfully {
		log.Fatal("unsuccessfully logged")
	}

	// setting up local repo (main db in app mode, temporary cache in org mode)
	cachePath := utils.GetAppCachePath()
	localRepo, err := database.NewRepo(&database.Config{
		Driver:       "sqlite",
		DataSource:   cachePath,
		SchemaString: db.SqliteSchema,
	})
	if err != nil {
		log.Printf("could not init cache repo (sqlite): %v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := localRepo.Log(ctx, []*domain.LogEntry{entry}); err != nil {
		log.Printf("error: could not write to local cache: %v", err)
	}

	if os.Getenv("APP_MODE") == "org" {
		var wg sync.WaitGroup
		wg.Add(1)

		go func() { // asynchronously flush cache and send to remote repo
			defer wg.Done()
			bgCtx, bgCancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer bgCancel()

			serverAddr := utils.GetEnvOrDefault("API_HOST_PORT", "localhost:9090")
			conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return // silently fail if server is offline leaving logs in cache
			}
			defer conn.Close()

			remoteRepo := grpcAdapter.NewClientAdapter(conn)
			multiRepo := database.NewMultiRepo(localRepo, remoteRepo)

			if _, err := multiRepo.FlushCache(bgCtx); err != nil {
				log.Printf("background flush failed: %v", err)
			}
		}()
		wg.Wait()
	}

	if *jsonMode {
		homeDir, _ := os.UserHomeDir()
		utils.LogJSON(grpcAdapter.LogEntryToProto(entry), utils.GetProjectRoot(homeDir))
	}
}
