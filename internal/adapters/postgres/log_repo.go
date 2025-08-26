package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/WillRabalais04/terminalLog/internal/core/domain"

	_ "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repository struct {
	db *sql.DB
}

func InitDB(dbURL string) (*sql.DB, error) {
	// func InitDB(host string, port int, user, password, dbname string) (*sql.DB, error) {
	// dsn := fmt.Sprintf(dbURL)
	// , host, port, user, password, dbname)

	db, err := sql.Open("pgx", dbURL) // pgx driver faster than pg driver
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}

func NewRepository(db *sql.DB) (*Repository, error) {
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

	return &Repository{db: db}, nil
}

func (r *Repository) Save(log domain.LogEntry) error {
	return nil
}

func (r *Repository) Get(id int) (domain.LogEntry, error) {
	return domain.LogEntry{}, nil
}
func (r *Repository) List() ([]domain.LogEntry, error) {
	return nil, nil
}
