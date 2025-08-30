package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/google/uuid"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

type LogRepo struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

type Config struct {
	Driver       string
	DataSource   string
	SchemaString string
}

func InitDB(driver, dataSource string) (*sql.DB, error) {
	db, err := sql.Open(driver, dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	switch driver {
	case "pgx":
		db.SetMaxOpenConns(25) // arbitrary should think about more later
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(5 * time.Minute)
	case "sqlite3":
		db.SetMaxOpenConns(1)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	return db, nil
}

func NewRepo(cfg *Config) (*LogRepo, error) {
	db, err := InitDB(cfg.Driver, cfg.DataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %v", err)
	}

	if _, err := db.Exec(cfg.SchemaString); err != nil {
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	var placeholder sq.PlaceholderFormat
	if cfg.Driver == "pgx" {
		placeholder = sq.Dollar
	}
	if cfg.Driver == "sqlite3" {
		placeholder = sq.Question
	}

	return &LogRepo{db: db, sb: sq.StatementBuilder.PlaceholderFormat(placeholder)}, nil
}

func (r *LogRepo) Log(ctx context.Context, entry *domain.LogEntry) error {
	entry.EventID = uuid.New().String()
	query := r.sb.Insert("logs").
		Columns(
			"event_id", "command", "exit_code", "ts", "shell_pid", "shell_uptime", "cwd",
			"prev_cwd", "user_name", "euid", "term", "hostname", "ssh_client",
			"tty", "git_repo", "git_repo_root", "git_branch", "git_commit",
			"git_status", "logged_successfully",
		).
		Values(
			entry.EventID,
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

func (r *LogRepo) Get(ctx context.Context, id int) (domain.LogEntry, error) {
	return domain.LogEntry{}, nil
}
func (r *LogRepo) List(ctx context.Context) ([]domain.LogEntry, error) {
	return nil, nil
}

func (r *LogRepo) Delete(ctx context.Context, id string) error {
	query := r.sb.Delete("logs").Where(sq.Eq{"event_id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	return nil
}
