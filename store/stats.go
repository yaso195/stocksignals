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
		"(signal_id, deposits, withdrawals, funds, balance, equity, profit, gain, drawdown, stats_time) "+
		"VALUES (:signal_id, :deposits, :withdrawals, :funds, :balance, :equity, :profit, :gain, :drawdown, :stats_time)",
		&stats)
	if err != nil {
		return fmt.Errorf("failed to insert stats : %s", err)
	}

	return nil
}

func insertStats(tx *sqlx.Tx, stats *model.Stats, profit float64, holdings []model.Holding) error {
	if tx == nil {
		return fmt.Errorf("given transaction is nil")
	}

	if stats.SignalID == 0 {
		return fmt.Errorf("stats cannot have signal ID 0")
	}

	if err := updateStats(stats, profit, holdings); err != nil {
		return fmt.Errorf("failed to update stats : %s", err)
	}

	stats.Time = time.Now().Unix()

	_, err := tx.NamedExec("INSERT INTO stats "+
		"(signal_id, deposits, withdrawals, funds, balance, equity, profit, gain, drawdown, stats_time) "+
		"VALUES (:signal_id, :deposits, :withdrawals, :funds, :balance, :equity, :profit, :gain, :drawdown, :stats_time)",
		&stats)
	if err != nil {
		return fmt.Errorf("failed to insert stats fund : %s", err)
	}

	return nil
}

// GetLatestStats reads the stats from the database based on the given signal id
func GetLatestStats(signalID int) (*model.Stats, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	var result model.Stats
	query := fmt.Sprintf("SELECT * FROM stats WHERE signal_id = %d ORDER BY stats_time DESC LIMIT 1", signalID)
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
	query := fmt.Sprintf("SELECT * FROM stats WHERE signal_id = %d ORDER BY stats_time DESC", signalID)

	if err := db.Select(&results, query); err != nil {
		return nil, fmt.Errorf("error reading stats: %q", err)
	}

	return results, nil
}

func updateStats(stats *model.Stats, profit float64, holdings []model.Holding) error {
	var stocks []string
	var totalStockBalance, totalStockEquity float64
	if len(holdings) > 0 {
		for _, holding := range holdings {
			stocks = append(stocks, holding.Code)
		}

		prices, err := stockapi.GetBidPrices(stocks)
		if err != nil {
			return err
		}

		for i, holding := range holdings {
			totalStockBalance += holding.Price * float64(holding.NumShares)
			totalStockEquity += prices[i] * float64(holding.NumShares)
		}
	}

	stats.Balance = totalStockBalance + stats.Funds
	stats.Equity = totalStockEquity + stats.Funds

	if stats.Equity < stats.Balance {
		drawdown := (stats.Balance - stats.Equity) * 100.0 / stats.Balance

		if drawdown > stats.Drawdown {
			stats.Drawdown = drawdown
		}
	}

	if profit != 0 {
		gain := (profit * 100.0 / stats.Balance)
		stats.Gain += gain
		stats.Profit += profit
	}

	return nil

}
