package store

import (
	"fmt"

	"github.com/heroku/stocksignals/model"
)

func createStats(signalID int) error {
	stats := model.Stats{ID: signalID}
	tx := db.MustBegin()
	_, err := tx.NamedExec("INSERT INTO stats (id, deposits, withdraws, funds, profits, num_trades)"+
		" VALUES (:id, :deposits, :withdraws, :funds, :profits, :num_trades)", &stats)
	if err != nil {
		return fmt.Errorf("failed to insert stats : %s", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete stats registration: %s", err)
	}
	return nil
}

func updateStatsFund(amount float64, tradeProfit bool, stats model.Stats) error {
	if stats.ID == 0 {
		return fmt.Errorf("funds cannot be updated for stats with ID 0")
	}

	if stats.Funds+amount < 0 {
		return fmt.Errorf("funds are not available to update")
	}

	stats.Funds += amount
	tx := db.MustBegin()

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

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete fund updating: %s", err)
	}

	return nil
}
