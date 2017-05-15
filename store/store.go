package store

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

const (
	DEFAULT_SIGNAL_FIELD  = "price"
	DEFAULT_ORDER_FIELD   = "order_time"
	DEFAULT_HOLDING_FIELD = "num_shares"
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

func Disconnect() error {
	if db == nil {
		return nil
	}

	return db.Close()
}
