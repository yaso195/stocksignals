package store

import (
	"database/sql"
	"fmt"

	"github.com/heroku/stocksignals/model"
	"github.com/jmoiron/sqlx"
)

func createStats(signalID int) error {
	stats := model.Stats{ID: signalID}
	tx := db.MustBegin()
	_, err := tx.NamedExec("INSERT INTO stats (id, deposits, withdrawals, funds, profit, num_trades)"+
		" VALUES (:id, :deposits, :withdrawals, :funds, :profit, :num_trades)", &stats)
	if err != nil {
		return fmt.Errorf("failed to insert stats : %s", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete stats registration: %s", err)
	}
	return nil
}

func updateStatsFund(tx *sqlx.Tx, amount float64, tradeProfit bool, stats model.Stats) error {
	if tx == nil {
		return fmt.Errorf("given transaction in update stats fund is nil")
	}

	if stats.ID == 0 {
		return fmt.Errorf("funds cannot be updated for stats with ID 0")
	}

	if stats.Funds+amount < 0 {
		return fmt.Errorf("funds are not available to update")
	}

	stats.Funds += amount

	if !tradeProfit {
		switch {
		case amount < 0:
			stats.Withdrawals += (-1) * amount
		case amount > 0:
			stats.Deposits += amount
		}
	}

	_, err := tx.NamedExec("UPDATE stats SET"+
		" funds = :funds, deposits = :deposits, withdrawals = :withdrawals WHERE id = :id", &stats)
	if err != nil {
		return fmt.Errorf("failed to update stats fund : %s", err)
	}

	return nil
}

// GetStats reads the stats from the database based on the given signal id
func GetStats(signalID int) (model.Stats, error) {
	if db == nil {
		return model.Stats{}, fmt.Errorf("no connection is created to the database")
	}

	var result model.Stats
	query := fmt.Sprintf("SELECT * FROM stats WHERE id = %d", signalID)
	err := db.Get(&result, query)
	if err == sql.ErrNoRows {
		return model.Stats{}, nil
	}

	if err != nil {
		return model.Stats{}, fmt.Errorf("error reading stats: %q", err)
	}

	return result, nil
}
