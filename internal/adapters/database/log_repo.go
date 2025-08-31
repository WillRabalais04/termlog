package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/WillRabalais04/terminalLog/internal/core/ports"
	"github.com/google/uuid"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

var logColumns = []string{
	"event_id", "command", "exit_code", "ts", "shell_pid", "shell_uptime", "cwd",
	"prev_cwd", "user_name", "euid", "term", "hostname", "ssh_client",
	"tty", "git_repo", "git_repo_root", "git_branch", "git_commit",
	"git_status", "logged_successfully",
}

var allowedOrderings = map[string]struct{}{
	"event_id":           struct{}{}, // why are these commented out?
	"exit_code":          struct{}{},
	"ts":                 struct{}{},
	"shell_pid":          struct{}{},
	"shell_uptime":       struct{}{},
	"cwd":                struct{}{},
	"prev_cwd":           struct{}{},
	"user_name":          struct{}{},
	"euid":               struct{}{},
	"term":               struct{}{},
	"hostname":           struct{}{},
	"ssh_client":         struct{}{},
	"tty":                struct{}{},
	"git_repo":           struct{}{},
	"git_repo_root":      struct{}{},
	"git_branch":         struct{}{},
	"git_commit":         struct{}{},
	"git_status":         struct{}{},
	"logged_succesfully": struct{}{},
}

var defaultOrdering = "ts"

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
		Columns(logColumns...).
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

// switch to calling list and with a filter that has event_id=id, user access and limit = 1
func (r *LogRepo) Get(ctx context.Context, id int) (*domain.LogEntry, error) {
	query := r.sb.Select("logs").Where(sq.Eq{"event_id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build get query: %w", err)
	}
	row := r.db.QueryRowContext(ctx, sqlStr, args...)

	return scanLogEntry(row)
}

func (r *LogRepo) List(ctx context.Context, filters *ports.LogQuery) ([]*domain.LogEntry, error) {
	query := sq.StatementBuilderType(r.sb.Select(logColumns...).From("logs"))
	query = applyFilters(query, filters)
	selectQuery := sq.SelectBuilder(query)

	orderBy, orderDir := validateOrdering(filters.OrderBy)
	selectQuery = selectQuery.OrderBy(fmt.Sprintf("%s %s", orderBy, orderDir))

	if filters.Limit > 0 {
		selectQuery = selectQuery.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		selectQuery = selectQuery.Offset(filters.Offset)
	}

	sqlStr, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build list query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute list query: %w", err)
	}
	defer rows.Close()

	var entries []*domain.LogEntry
	for rows.Next() {
		entry, err := scanLogEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return entries, nil
}

func (r *LogRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteMultiple(ctx, &ports.LogQuery{EventID: &id})
}

func (r *LogRepo) DeleteMultiple(ctx context.Context, filters *ports.LogQuery) error {
	query := sq.StatementBuilderType(r.sb.Delete("logs"))
	query = applyFilters(query, filters)

	sqlStr, args, err := (sq.DeleteBuilder(query)).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}
	return nil
}

// add pagination and user access filtering down the line
func applyFilters(builder sq.StatementBuilderType, filters *ports.LogQuery) sq.StatementBuilderType {
	if filters.EventID != nil {
		builder = builder.Where(sq.Eq{"event_id": *filters.EventID})
	}
	if filters.Command != nil {
		builder = builder.Where(sq.Like{"command": *filters.Command})
	}
	if filters.ExitCode != nil {
		builder = builder.Where(sq.Eq{"exit_code": *filters.ExitCode})
	}
	// if filters.Limit != nil { // implement
	// 	builder = builder.Where(sq.Eq{"logged_successfully": *filters.LoggedSuccesfully})
	// }
	if filters.Timestamp != nil {
		builder = builder.Where(sq.Eq{"ts": *filters.Timestamp})
	}
	if filters.ShellPID != nil {
		builder = builder.Where(sq.Eq{"shell_pid": *filters.ShellPID})
	}
	if filters.ShellUptime != nil {
		builder = builder.Where(sq.Eq{"shell_uptime": *filters.ShellUptime})
	}
	if filters.WorkingDirectory != nil {
		builder = builder.Where(sq.Eq{"cwd": *filters.WorkingDirectory})
	}
	if filters.PrevWorkingDirectory != nil {
		builder = builder.Where(sq.Eq{"prev_cwd": *filters.PrevWorkingDirectory})
	}
	if filters.User != nil {
		builder = builder.Where(sq.Eq{"user_name": *filters.User})
	}
	if filters.EUID != nil {
		builder = builder.Where(sq.Eq{"euid": *filters.EUID})
	}
	if filters.Term != nil {
		builder = builder.Where(sq.Eq{"term": *filters.Term})
	}
	if filters.Hostname != nil {
		builder = builder.Where(sq.Eq{"hostname": *filters.Hostname})
	}
	if filters.SSHClient != nil {
		builder = builder.Where(sq.Eq{"ssh_client": *filters.SSHClient})
	}
	if filters.TTY != nil {
		builder = builder.Where(sq.Eq{"tty": *filters.TTY})
	}
	if filters.GitRepo != nil {
		builder = builder.Where(sq.Eq{"git_repo": *filters.GitRepo})
	}
	if filters.GitRepoRoot != nil {
		builder = builder.Where(sq.Eq{"git_repo_root": *filters.GitRepoRoot})
	}
	if filters.Branch != nil {
		builder = builder.Where(sq.Eq{"branch": *filters.Branch})
	}
	if filters.Commit != nil {
		builder = builder.Where(sq.Eq{"git_commit": *filters.Commit})
	}
	if filters.LoggedSuccesfully != nil {
		builder = builder.Where(sq.Eq{"logged_successfully": *filters.LoggedSuccesfully})
	}

	return builder
}

func scanLogEntry(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.LogEntry, error) {
	var entry domain.LogEntry
	var ts time.Time

	err := scanner.Scan(
		&entry.EventID,
		&entry.Command,
		&entry.ExitCode,
		&ts,
		&entry.Shell_PID,
		&entry.ShellUptime,
		&entry.WorkingDirectory,
		&entry.PrevWorkingDirectory,
		&entry.User,
		&entry.EUID,
		&entry.Term,
		&entry.Hostname,
		&entry.SSHClient,
		&entry.TTY,
		&entry.IsGitRepo,
		&entry.GitRepoRoot,
		&entry.GitBranch,
		&entry.GitCommit,
		&entry.GitStatus,
		&entry.LoggedSuccessfully,
	)

	if err != nil {
		return nil, err
	}

	entry.Timestamp = ts.Unix()

	return &entry, nil
}

func validateOrdering(ordering *string) (string, string) {
	orderDir := "DESC"
	validatedOrdering := defaultOrdering

	if ordering != nil {
		derefedOrdering := strings.ToLower(*ordering)
		if derefedOrdering[0] == '-' {
			derefedOrdering = derefedOrdering[1:]
			orderDir = "ASC"
		}
		if _, ok := allowedOrderings[derefedOrdering]; ok {
			validatedOrdering = derefedOrdering
		}
	}

	return validatedOrdering, orderDir
}
