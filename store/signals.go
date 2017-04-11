package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/heroku/stocksignals/model"
)

// Reads the signals from the database and orders them based on the given field
func GetSignals(field string, descend bool) ([]model.Signal, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	if field == "" {
		field = DEFAULT_SIGNAL_ORDER_FIELD
	}

	order := "DESC"
	if !descend {
		order = "ASC"
	}

	var results []model.Signal
	err := db.Select(&results, fmt.Sprintf("SELECT * FROM signals ORDER BY %s %s", field, order))
	if err != nil {
		return nil, fmt.Errorf("error reading signals: %q", err)
	}

	return results, nil
}

// RegisterSignal registers the given signal to the database
func RegisterSignal(signal model.Signal) error {
	if db == nil {
		return fmt.Errorf("no connection is created to the database")
	}

	if signal.Name == "" {
		return fmt.Errorf("signal name cannot be empty")
	}

	tempName := strings.TrimSpace(strings.ToLower(signal.Name))
	if signal.Price <= 0 {
		return fmt.Errorf("price cannot be less than or equal to 0")
	}

	var result model.Signal
	err := db.Get(&result, fmt.Sprintf("SELECT * FROM signals WHERE lower(name)='%s'", tempName))
	if err == sql.ErrNoRows {
		_, errRegister := db.NamedExec("INSERT INTO signals (name, description, num_subscribers, num_trades, price, growth)"+
			" VALUES (:name, :description, :num_subscribers, :num_trades, :price, :growth)", &signal)
		if errRegister != nil {
			return fmt.Errorf("error registering signal with name %s: %q", signal.Name, err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading signal with name %s: %q", signal.Name, err)
	}

	return fmt.Errorf("signal already exists with name %s", signal.Name)
}

// Reads the signals from the database by ID, returns empty ID if it cannot find it
func GetSignalByID(id int) (*model.Signal, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	if id < 0 {
		return nil, fmt.Errorf("invalid signal id")
	}
	var result model.Signal
	err := db.Get(&result, fmt.Sprintf("SELECT * FROM signals WHERE id=%d", id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("signal with id %d does not exist.", id)
	}

	if err != nil {
		return nil, fmt.Errorf("error reading signal with id %d: %q", id, err)
	}

	return &result, nil
}
