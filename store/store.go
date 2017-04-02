package store

import (
	"fmt"
	"os"

	"github.com/heroku/stocksignals/model"
	"github.com/jmoiron/sqlx"
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

// Reads the signals from the database and orders them based on their growth
func GetSignals() ([]model.Signal, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	var results []model.Signal
	err := db.Select(&results, "SELECT * FROM signals ORDER_BY growth DESC")
	if err != nil {
		return nil, fmt.Errorf("error reading signals: %q", err)
	}

	return results, nil
}
