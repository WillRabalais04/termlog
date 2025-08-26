package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type LogRepo struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func InitDB(dbURL string) (*sql.DB, error) {

	db, err := sql.Open("pgx", dbURL) // pgx driver faster than pg driver
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}

func NewRepository(db *sql.DB) (*LogRepo, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	log.Println("Successfully connected to DB")

	// migration
	if schema, err := os.ReadFile("db/schema.sql"); err == nil {
		if _, err := db.Exec(string(schema)); err != nil {
			log.Printf("Failed to execute schema: %v", err)
		} else {
			log.Println("Schema applied successfully")
		}
	}

	return &LogRepo{db: db, sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}, nil
}

func (r *LogRepo) Save(entry domain.LogEntry) error {
	query := r.sb.Insert("logs").
		Columns(
			"command", "exit_code", "ts", "shell_pid", "shell_uptime", "cwd",
			"prev_cwd", "user_name", "euid", "term", "hostname", "ssh_client",
			"tty", "git_repo", "git_repo_root", "git_branch", "git_commit",
			"git_status", "logged_successfully",
		).
		Values(
			entry.Command,
			entry.ExitCode,
			time.Unix(entry.Timestamp, 0),
			entry.Shell_PID,
			entry.ShellUptime,
			entry.WorkingDirectory,
			entry.PrevWorkingDirectory,
			entry.User,
			entry.EUID,
			entry.Term,
			entry.Hostname,
			entry.SSHClient,
			entry.TTY,
			entry.IsGitRepo,
			entry.GitRepoRoot,
			entry.GitBranch,
			entry.GitCommit,
			entry.GitStatus,
			entry.LoggedSuccessfully,
		)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(sqlStr, args...)
	return err
}

func (r *LogRepo) Get(id int) (domain.LogEntry, error) {
	return domain.LogEntry{}, nil
}
func (r *LogRepo) List() ([]domain.LogEntry, error) {
	return nil, nil
}
