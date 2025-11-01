package database

import (
	"database/sql"
	"os"
)

func SetupDatabase() (*sql.DB, error) {
	dbURL := getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/corroboros?sslmode=disable")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	// run migrations or initial setup if needed
	return db, nil
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
