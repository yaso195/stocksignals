package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/heroku/stocksignals/model"
	"github.com/heroku/stocksignals/stockapi"

	"github.com/jmoiron/sqlx"
)

func createNewStats(tx *sqlx.Tx, signalID int) error {
	stats := model.Stats{SignalID: signalID, Time: time.Now().Unix()}

	_, err := tx.NamedExec("INSERT INTO stats "+
		"(signal_id, deposits, withdrawals, funds, balance, equity, profit, growth, drawdown, stats_time) "+
		"VALUES (:signal_id, :deposits, :withdrawals, :funds, :balance, :equity, :profit, :growth, :drawdown, :stats_time)",
		&stats)
	if err != nil {
		return fmt.Errorf("failed to insert stats : %s", err)
	}

	return nil
}

func insertStats(tx *sqlx.Tx, stats *model.Stats, profit, previousBalance float64, holdings []model.Holding, pastStats bool) error {
	if tx == nil {
		return fmt.Errorf("given transaction is nil")
	}

	if stats.SignalID == 0 {
		return fmt.Errorf("stats cannot have signal ID 0")
	}

	if err := updateStats(stats, profit, previousBalance, holdings, pastStats); err != nil {
		return fmt.Errorf("failed to update stats : %s", err)
	}

	if stats.Time == 0 {
		stats.Time = time.Now().Unix()
	}

	_, err := tx.NamedExec("INSERT INTO stats "+
		"(signal_id, deposits, withdrawals, funds, balance, equity, profit, growth, drawdown, stats_time) "+
		"VALUES (:signal_id, :deposits, :withdrawals, :funds, :balance, :equity, :profit, :growth, :drawdown, :stats_time)",
		&stats)
	if err != nil {
		return fmt.Errorf("failed to insert stats %v : %s", stats, err)
	}

	return nil
}

// GetLatestStats reads the stats from the database based on the given signal id
func GetLatestStats(signalID int) (*model.Stats, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	var result model.Stats
	query := fmt.Sprintf("SELECT * FROM stats WHERE signal_id = %d ORDER BY (stats_time, id) DESC LIMIT 1", signalID)
	err := db.Get(&result, query)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error reading stats: %q", err)
	}

	return &result, nil
}

// GetAllStats reads the stats from the database based on the given signal id
func GetAllStats(signalID int) ([]model.Stats, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	var results []model.Stats
	query := fmt.Sprintf("SELECT * FROM stats WHERE signal_id = %d ORDER BY (stats_time, id) DESC", signalID)

	if err := db.Select(&results, query); err != nil {
		return nil, fmt.Errorf("error reading stats: %q", err)
	}

	return results, nil
}

func updateStats(stats *model.Stats, profit, previousBalance float64, holdings []model.Holding, pastStats bool) error {
	var stocks []string
	var totalStockBalance, totalStockEquity float64
	if len(holdings) > 0 {
		for _, holding := range holdings {
			stocks = append(stocks, holding.Code)
		}

		var prices []float64
		var err error

		if !pastStats {
			prices, err = stockapi.GetBidPrices(stocks)
			if err != nil {
				return err
			}
		}

		for i, holding := range holdings {
			totalStockBalance += holding.Price * float64(holding.NumShares)
			if !pastStats {
				totalStockEquity += prices[i] * float64(holding.NumShares)
			}
		}

		if pastStats {
			totalStockEquity = totalStockBalance
		}
	}

	stats.Balance = totalStockBalance + stats.Funds
	stats.Equity = totalStockEquity + stats.Funds
	if stats.Balance != 0 {
		stats.Drawdown = (stats.Balance - stats.Equity) * 100.0 / stats.Balance
	}

	if previousBalance != 0 {
		gain := (profit * 100.0 / previousBalance)
		stats.Growth += gain
	}
	stats.Profit += profit

	return nil
}

func deleteStatsBySignalID(signalID int, tx *sqlx.Tx) error {
	if tx == nil {
		return fmt.Errorf("given transaction is nil")
	}

	_, err := tx.Exec(fmt.Sprintf("DELETE FROM stats WHERE signal_id = %d", signalID))
	if err != nil {
		return fmt.Errorf("failed to delete stats from store : %s", err)
	}

	return nil
}

func SaveStats(signalID int) error {
	tx := db.MustBegin()

	stats, err := GetLatestStats(signalID)
	if err != nil {
		return fmt.Errorf("failed to get latest stats for signal %d: %s", signalID, err)
	}

	// Zero the stats time so that new stat would come with new time
	stats.Time = 0

	holdings, err := GetHoldingsBySignalID(signalID, "", true)
	if err != nil {
		return fmt.Errorf("failed to get holdings for signal %d: %s", signalID, err)
	}

	previousBalance := getStockBalance(holdings) + stats.Funds

	if err = insertStats(tx, stats, 0, previousBalance, holdings, false); err != nil {
		return fmt.Errorf("failed to insert new stats for signal %d: %s", signalID, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit to new computed stats for signal %d : %s", signalID, err)
	}

	return nil
}
