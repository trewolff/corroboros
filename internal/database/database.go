package database

import (
	"database/sql"

	"github.com/trewolff/corroboros/internal/config"
)

func SetupDatabase(cfg config.Config) (*sql.DB, error) {
	dbURL := cfg.DBConnectionString
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	// run migrations or initial setup if needed
	return db, nil
}
