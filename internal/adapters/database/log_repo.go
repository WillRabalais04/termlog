package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	"github.com/google/uuid"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

var logColumns = []string{
	"event_id", "command", "exit_code", "ts", "shell_pid", "shell_uptime", "cwd",
	"prev_cwd", "user_name", "euid", "term", "hostname", "ssh_client",
	"tty", "git_repo", "git_repo_root", "git_branch", "git_commit",
	"git_status", "logged_successfully",
}

var columnMetadata = map[string]struct {
	Type        interface{}
	IsOrderable bool
	IsExact     bool
	IsFuzzy     bool
}{
	"event_id":            {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: false},
	"command":             {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"exit_code":           {Type: int32(0), IsOrderable: true, IsExact: true, IsFuzzy: false},
	"ts":                  {Type: int64(0), IsOrderable: true, IsExact: true, IsFuzzy: false},
	"shell_pid":           {Type: int32(0), IsOrderable: true, IsExact: true, IsFuzzy: false},
	"shell_uptime":        {Type: int64(0), IsOrderable: true, IsExact: true, IsFuzzy: false},
	"cwd":                 {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"prev_cwd":            {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"user_name":           {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"euid":                {Type: int32(0), IsOrderable: true, IsExact: true, IsFuzzy: false},
	"term":                {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"hostname":            {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"ssh_client":          {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"tty":                 {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"git_repo":            {Type: false, IsOrderable: true, IsExact: true, IsFuzzy: false},
	"git_repo_root":       {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"git_branch":          {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"git_commit":          {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"git_status":          {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"logged_successfully": {Type: false, IsOrderable: true, IsExact: true, IsFuzzy: false},
}

var allowedOrderings map[string]struct{}
var defaultOrdering = "ts"

type Config struct {
	Driver       string
	DataSource   string
	SchemaString string
}

type LogRepo struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func init() {
	allowedOrderings = make(map[string]struct{}, len(logColumns))
	for _, col := range logColumns {
		allowedOrderings[col] = struct{}{}
	}
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
	switch cfg.Driver {
	case "pgx":
		placeholder = sq.Dollar
	case "sqlite":
		placeholder = sq.Question
	default:
		return nil, fmt.Errorf("invalid db driver name (should pgx or sqlite3)")
	}

	return &LogRepo{db: db, sb: sq.StatementBuilder.PlaceholderFormat(placeholder)}, nil
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
	case "sqlite":
		db.SetMaxOpenConns(1)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	return db, nil
}

func (r *LogRepo) Log(ctx context.Context, entries []*domain.LogEntry) error { // switched to batched logging for efficiency in pushing cache
	if len(entries) == 0 {
		return nil
	}

	query := r.sb.Insert("logs").Columns(logColumns...)

	for _, entry := range entries {
		if entry.EventID == "" {
			entry.EventID = uuid.New().String()
		}
		query = query.Values(
			entry.EventID, entry.Command, entry.ExitCode, entry.Timestamp,
			entry.Shell_PID, entry.ShellUptime, entry.WorkingDirectory, entry.PrevWorkingDirectory,
			entry.User, entry.EUID, entry.Term, entry.Hostname,
			entry.SSHClient, entry.TTY, entry.GitRepo, entry.GitRepoRoot,
			entry.GitBranch, entry.GitCommit, entry.GitStatus, entry.LoggedSuccessfully,
		)
	}

	query = query.Suffix("ON CONFLICT(event_id) DO NOTHING") // prevent duplicates (idempotent)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build bulk insert query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return err
}

func (r *LogRepo) Get(ctx context.Context, id string) (*domain.LogEntry, error) {
	filter := domain.NewFilterBuilder().
		AddFilterTerm("event_id", id).
		Build()

	entries, err := r.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get query: %w", err)
	}

	if len(entries) != 1 {
		return nil, fmt.Errorf("failed to get entry")
	}
	return entries[0], nil
}

func (r *LogRepo) List(ctx context.Context, filter *domain.LogFilter) ([]*domain.LogEntry, error) {
	query := sq.StatementBuilderType(r.sb.Select(logColumns...).From("logs"))
	query = applyFilters(query, filter)
	selectQuery := sq.SelectBuilder(query)
	orderBy, orderDir := validateOrdering(filter.OrderBy)
	selectQuery = selectQuery.OrderBy(fmt.Sprintf("%s %s", orderBy, orderDir))

	if filter.Limit > 0 {
		selectQuery = selectQuery.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		selectQuery = selectQuery.Offset(filter.Offset)
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

func (r *LogRepo) Delete(ctx context.Context, id string) (*domain.LogEntry, error) {
	filter := domain.NewFilterBuilder().
		AddFilterTerm("event_id", id).
		Build()
	deletedEntries, err := r.DeleteMultiple(ctx, filter)

	if err != nil {
		return nil, fmt.Errorf("database error during delete: %w", err)
	}
	if len(deletedEntries) == 0 {
		return nil, nil
	}
	if len(deletedEntries) > 1 {
		return nil, fmt.Errorf("consistency error: multiple entries deleted for id %s", id)
	}

	return deletedEntries[0], nil
}

func (r *LogRepo) DeleteMultiple(ctx context.Context, filter *domain.LogFilter) ([]*domain.LogEntry, error) {
	query := sq.StatementBuilderType(r.sb.Delete("logs"))
	query = applyFilters(query, filter)
	sqlStr, args, err := (sq.DeleteBuilder(query)).Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build delete query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute delete query: %w", err)
	}
	defer rows.Close()

	var deletedEntries []*domain.LogEntry
	for rows.Next() {
		entry, err := scanLogEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}
		deletedEntries = append(deletedEntries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return deletedEntries, nil
}

func applyFilters(builder sq.StatementBuilderType, filter *domain.LogFilter) sq.StatementBuilderType {
	builder = applyFilterTerms(builder, filter.FilterTerms, filter.FilterMode)
	builder = applySearchTerms(builder, filter.SearchTerms, filter.SearchMode)
	// add user permissions filter later
	if filter.StartTime != nil {
		builder = builder.Where(sq.GtOrEq{"ts": *filter.StartTime})
	}
	if filter.EndTime != nil {
		builder = builder.Where(sq.LtOrEq{"ts": *filter.EndTime})
	}
	return builder
}

func applyFilterTerms(builder sq.StatementBuilderType, filterTerms map[string]domain.FilterValues, mode domain.Mode) sq.StatementBuilderType {
	if len(filterTerms) == 0 {
		return builder
	}
	var allFieldConditions []sq.Sqlizer
	for field, values := range filterTerms {
		metadata, ok := columnMetadata[field]
		if !ok || !metadata.IsExact {
			continue
		}
		var fieldConditions []sq.Sqlizer
		for _, val := range values.Values {
			typedValue, err := convertValue(val, metadata.Type)
			if err != nil {
				fmt.Printf("error converting value for field '%s': %v\n", field, err)
				continue
			}
			fieldConditions = append(fieldConditions, sq.Eq{field: typedValue})
		}
		if len(fieldConditions) > 0 {
			allFieldConditions = append(allFieldConditions, sq.Or(fieldConditions))
		}
	}
	if len(allFieldConditions) > 0 {
		if mode == domain.AND {
			builder = builder.Where(sq.And(allFieldConditions))
		} else {
			builder = builder.Where(sq.Or(allFieldConditions))
		}
	}
	return builder
}

func applySearchTerms(builder sq.StatementBuilderType, searchTerms map[string]domain.SearchValues, mode domain.Mode) sq.StatementBuilderType {
	if len(searchTerms) == 0 {
		return builder
	}
	var allFieldConditions []sq.Sqlizer
	for field, values := range searchTerms {
		metadata, ok := columnMetadata[field]
		if !ok || !metadata.IsFuzzy {
			continue
		}
		var fieldConditions []sq.Sqlizer
		for _, val := range values.Values {
			likeTerm := "%" + val + "%"
			condition := sq.Expr("LOWER("+field+") LIKE LOWER(?)", likeTerm) // lower case index for faster fuzzy search (also sqlite doesn't support ILIKE)
			fieldConditions = append(fieldConditions, condition)
		}
		if len(fieldConditions) > 0 {
			allFieldConditions = append(allFieldConditions, sq.Or(fieldConditions))
		}
	}
	if len(allFieldConditions) > 0 {
		if mode == domain.AND {
			builder = builder.Where(sq.And(allFieldConditions))
		} else {
			builder = builder.Where(sq.Or(allFieldConditions))
		}
	}
	return builder
}

func convertValue(val string, targetType interface{}) (interface{}, error) {
	switch targetType.(type) {
	case string:
		return val, nil
	case int32:
		i, err := strconv.ParseInt(val, 10, 32)
		return int32(i), err
	case int64:
		return strconv.ParseInt(val, 10, 64)
	case bool:
		return strconv.ParseBool(val)
	default:
		return nil, fmt.Errorf("unsupported type: %T", targetType)
	}
}

func validateOrdering(ordering *string) (string, string) {
	orderDir := "DESC"
	validatedOrdering := defaultOrdering

	if ordering != nil {
		derefedOrdering := strings.ToLower(*ordering)
		if strings.HasPrefix(derefedOrdering, "-") {
			derefedOrdering = strings.TrimPrefix(derefedOrdering, "-")
			orderDir = "ASC"
		}
		metadata, ok := columnMetadata[derefedOrdering]
		if ok && metadata.IsOrderable {
			validatedOrdering = derefedOrdering
		}
	}

	return validatedOrdering, orderDir
}

func scanLogEntry(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.LogEntry, error) {
	var entry domain.LogEntry
	err := scanner.Scan(
		&entry.EventID,
		&entry.Command,
		&entry.ExitCode,
		&entry.Timestamp,
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
		&entry.GitRepo,
		&entry.GitRepoRoot,
		&entry.GitBranch,
		&entry.GitCommit,
		&entry.GitStatus,
		&entry.LoggedSuccessfully,
	)

	if err != nil {
		return nil, err
	}

	return &entry, nil
}
