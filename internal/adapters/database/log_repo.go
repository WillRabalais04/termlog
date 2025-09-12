package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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
	"is_git_repo":         {Type: false, IsOrderable: true, IsExact: true, IsFuzzy: false},
	"git_repo_root":       {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"git_branch":          {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"git_commit":          {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"git_status":          {Type: "", IsOrderable: true, IsExact: true, IsFuzzy: true},
	"logged_successfully": {Type: false, IsOrderable: true, IsExact: true, IsFuzzy: false},
}

var allowedOrderings map[string]struct{}

func init() {
	allowedOrderings = make(map[string]struct{}, len(logColumns))
	for _, col := range logColumns {
		allowedOrderings[col] = struct{}{}
	}
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
	switch cfg.Driver {
	case "pgx":
		placeholder = sq.Dollar
	case "sqlite3":
		placeholder = sq.Question
	default:
		return nil, fmt.Errorf("invalid db driver name (should pgx or sqlite3)")
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
			entry.Timestamp,
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

// switch to calling list and with a filter that has event_id=id, user access and limit = 1 (if faster)
func (r *LogRepo) Get(ctx context.Context, id string) (*domain.LogEntry, error) {
	filter := &ports.LogFilter{
		FilterTerms: map[string]ports.FilterValues{
			"event_id": {Values: []string{id}},
		},
		FilterMode: ports.AND,
	}

	entries, err := r.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get query: %w", err)
	}

	return entries[0], nil
}

func (r *LogRepo) List(ctx context.Context, filter *ports.LogFilter) ([]*domain.LogEntry, error) {
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

func (r *LogRepo) Delete(ctx context.Context, id string) error {
	filter := &ports.LogFilter{
		FilterTerms: map[string]ports.FilterValues{
			"event_id": {Values: []string{id}},
		},
		FilterMode: ports.AND,
	}
	return r.DeleteMultiple(ctx, filter)
}

func (r *LogRepo) DeleteMultiple(ctx context.Context, filter *ports.LogFilter) error {
	query := sq.StatementBuilderType(r.sb.Delete("logs"))
	query = applyFilterTerms(query, filter.FilterTerms, filter.FilterMode) // use applyfilters maybe?

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

func applyFilters(builder sq.StatementBuilderType, filter *ports.LogFilter) sq.StatementBuilderType {
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

func applyFilterTerms(builder sq.StatementBuilderType, filterTerms map[string]ports.FilterValues, mode ports.Mode) sq.StatementBuilderType {
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
				fmt.Printf("Error converting value for field '%s': %v\n", field, err)
				continue
			}
			fieldConditions = append(fieldConditions, sq.Eq{field: typedValue})
		}
		if len(fieldConditions) > 0 {
			allFieldConditions = append(allFieldConditions, sq.Or(fieldConditions))
		}
	}
	if len(allFieldConditions) > 0 {
		if mode == ports.AND {
			builder = builder.Where(sq.And(allFieldConditions))
		} else {
			builder = builder.Where(sq.Or(allFieldConditions))
		}
	}
	return builder
}

func applySearchTerms(builder sq.StatementBuilderType, searchTerms map[string]ports.SearchValues, mode ports.Mode) sq.StatementBuilderType {
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
			fieldConditions = append(fieldConditions, sq.ILike{field: likeTerm})
		}
		if len(fieldConditions) > 0 {
			allFieldConditions = append(allFieldConditions, sq.Or(fieldConditions))
		}
	}
	if len(allFieldConditions) > 0 {
		if mode == ports.AND {
			builder = builder.Where(sq.And(allFieldConditions))
		} else {
			builder = builder.Where(sq.Or(allFieldConditions))
		}
	}
	return builder
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

	return &entry, nil
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
