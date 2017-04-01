package store

import (
	"database/sql"
	"fmt"
	"os"
)

var (
	db *sql.DB
)

func Connect() error {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("Error opening database: %q", err)
	}
	return nil
}
