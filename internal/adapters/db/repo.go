package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

type LogRepo struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

type Config struct {
	Driver     string
	DataSource string
	SchemaFile string
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
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to ping postgres: %w", err)
		}
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(5 * time.Minute)
		log.Println("successfully connected to postgres")
	case "sqlite3":
		db.SetMaxOpenConns(1)
		log.Println("successfully connected to sqlite")
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

	schema, err := os.ReadFile(cfg.SchemaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", cfg.SchemaFile, err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Printf("schema from %s applied successfully", cfg.SchemaFile)
	var placeholder sq.PlaceholderFormat

	if cfg.Driver == "pgx" {
		placeholder = sq.Dollar
	}
	if cfg.Driver == "sqlite3" {
		placeholder = sq.Question
	}

	return &LogRepo{db: db, sb: sq.StatementBuilder.PlaceholderFormat(placeholder)}, nil
}

func (r *LogRepo) Save(entry *domain.LogEntry) error {
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
