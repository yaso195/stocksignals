package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/heroku/stocksignals/model"

	"github.com/jmoiron/sqlx"
)

// Reads the signals from the database and orders them based on the given field
func GetSignals(field string, descend bool) ([]model.Signal, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	if field == "" {
		field = DEFAULT_SIGNAL_FIELD
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

// RegisterSignals registers the given signals to the database
func RegisterSignals(signals []model.Signal) error {
	if db == nil {
		return fmt.Errorf("no connection is created to the database")
	}

	tx := db.MustBegin()

	for _, signal := range signals {
		if err := registerSignal(signal, tx); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete signals registration : %s", err)
	}
	return nil
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

// DeleteSignalsByID deletes the given signals from the database
// It cleans up all the orders, stats and holding for this signal.
func DeleteSignalsByID(ids []int) error {
	tx := db.MustBegin()

	var err error
	for _, id := range ids {
		_, err = GetSignalByID(id)
		if err != nil {
			return err
		}

		if err = deleteOrdersBySignalID(id, tx); err != nil {
			return err
		}

		if err = deleteStatsBySignalID(id, tx); err != nil {
			return err
		}

		if err = deleteHoldingsBySignalID(id, tx); err != nil {
			return err
		}

		_, err = tx.Exec(fmt.Sprintf("DELETE FROM signals WHERE id = %d", id))
		if err != nil {
			return fmt.Errorf("failed to delete signal from store : %s", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete signal deletion : %s", err)
	}

	return nil
}

func registerSignal(signal model.Signal, tx *sqlx.Tx) error {
	if tx == nil {
		return fmt.Errorf("given transaction is nil")
	}

	if signal.Name == "" {
		return fmt.Errorf("signal name cannot be empty")
	}

	tempName := strings.TrimSpace(strings.ToLower(signal.Name))
	if signal.Price <= 0 {
		return fmt.Errorf("price cannot be less than or equal to 0")
	}

	var result model.Signal
	err := tx.Get(&result, fmt.Sprintf("SELECT * FROM signals WHERE lower(name)='%s'", tempName))
	if err == sql.ErrNoRows {
		var id int
		errRegister := tx.QueryRow("INSERT INTO signals (name, description, num_subscribers, price, num_trades, "+
			"first_trade_time, last_trade_time) VALUES ($1, $2, $3, $4, $5, $6, $7) returning id",
			signal.Name, signal.Description, signal.NumSubscribers, signal.Price, signal.NumTrades, signal.FirstTradeTime, signal.LastTradeTime).Scan(&id)
		if errRegister != nil {
			return fmt.Errorf("error registering signal with name %s: %q", signal.Name, err)
		}

		if err = createNewStats(tx, id); err != nil {
			return fmt.Errorf("failed to create new stats : %s", err)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading signal with name %s: %q", signal.Name, err)
	}

	return fmt.Errorf("signal already exists with name %s", signal.Name)
}
