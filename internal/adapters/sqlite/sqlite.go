// internal/adapters/sqlite/sqlite.go
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/WillRabalais04/terminalLog/internal/core/domain"
	_ "github.com/mattn/go-sqlite3"
)

type LogRepo struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func InitDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open local db: %w", err)
	}

	return db, nil
}

func NewRepository(db *sql.DB) (*LogRepo, error) {
	if schema, err := os.ReadFile("db/schema.sqlite.sql"); err == nil {
		if _, err := db.Exec(string(schema)); err != nil {
			log.Printf("Failed to execute schema: %v", err)
		} else {
			log.Println("SQLite schema applied successfully")
		}
	} else {
		log.Printf("No SQLite schema file found (skipping migration): %v", err)
	}

	return &LogRepo{db: db, sb: sq.StatementBuilder.PlaceholderFormat(sq.Question)}, nil
}

func (r *LogRepo) Save(ctx context.Context, entry *domain.LogEntry) error {

	return err
}

func (r *LogRepo) Get(id int) (domain.LogEntry, error) {
	return domain.LogEntry{}, nil
}
func (r *LogRepo) List() ([]domain.LogEntry, error) {
	return nil, nil
}

// func (r *LogRepo) Close() error {
// 	return r.db.Close()
// }
