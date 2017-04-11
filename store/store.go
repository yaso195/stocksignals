package store

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

const (
	DEFAULT_SIGNAL_ORDER_FIELD = "growth"
)

var (
	db *sqlx.DB
)

func Connect() error {
	var err error
	db, err = sqlx.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("Error opening database: %q %q", err, os.Getenv("DATABASE_URL"))
	}
	return nil
}
