package store

import (
	"fmt"

	"github.com/heroku/stocksignals/model"
	"github.com/jmoiron/sqlx"
)

//  GetOrdersBySignalID reads the orders from the database based on the given signal idand orders them based on the given field
func GetOrdersBySignalID(signalID int, field string, descend bool) ([]model.Order, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	if field == "" {
		field = DEFAULT_ORDER_FIELD
	}

	order := "DESC"
	if !descend {
		order = "ASC"
	}

	var results []model.Order
	err := db.Select(&results, fmt.Sprintf("SELECT * FROM orders WHERE signal_id = %d ORDER BY %s %s", signalID, field, order))
	if err != nil {
		return nil, fmt.Errorf("error reading orders: %q", err)
	}

	return results, nil
}

func RegisterOrder(order *model.Order) error {
	if order == nil {
		return fmt.Errorf("given order is nil")
	}

	signal, err := GetSignalByID(order.SignalID)
	if err != nil {
		return err
	}

	stats, err := GetLatestStats(order.SignalID)
	if err != nil {
		return err
	}

	holdings, err := GetHoldingsBySignalID(order.SignalID, "", true)
	if err != nil {
		return err
	}

	tx := db.MustBegin()

	var statsProfit float64
	switch order.Type {
	case model.DEPOSIT:
		err = executeDepositOrder(stats, order)
	case model.WITHDRAW:
		err = executeWithdrawOrder(stats, order)
	case model.BUY, model.ADD:
		loc := findHolding(order.Code, holdings)
		var holding *model.Holding
		if loc == -1 {
			holding = &model.Holding{}
		} else {
			holding = &holdings[loc]
		}

		err = executeBuyOrder(*signal, stats, order, holding, tx)

		if loc == -1 {
			holdings = append(holdings, *holding)
		} else {
			holdings[loc] = *holding
		}
	case model.SELL, model.REDUCE:
		loc := findHolding(order.Code, holdings)
		var holding *model.Holding
		if loc != -1 {
			holding = &holdings[loc]
		}

		err = executeSellOrder(*signal, stats, order, holding, tx)

		statsProfit = order.Profit
	default:
		err = fmt.Errorf("unknown order type")
	}

	if err != nil {
		return fmt.Errorf("failed to prepare %s order : %s", order.Type, err)
	}

	_, err = tx.NamedExec("INSERT INTO orders (signal_id, order_time, type, code, name, num_shares, price, profit)"+
		" VALUES (:signal_id, :order_time, :type, :code, :name, :num_shares, :price, :profit)", &order)
	if err != nil {
		return fmt.Errorf("failed to insert order : %s", err)
	}

	if err = insertStats(tx, stats, statsProfit, holdings); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete order registration : %s", err)
	}
	return nil
}

func executeDepositOrder(stats *model.Stats, order *model.Order) error {
	if stats.SignalID == 0 || stats.SignalID != order.SignalID {
		return fmt.Errorf("given stats is invalid")
	}

	stats.Funds += order.Profit
	stats.Deposits += order.Profit

	order.Code = ""
	order.Name = ""
	order.NumShares = 0
	order.Price = 0
	return nil
}

func executeWithdrawOrder(stats *model.Stats, order *model.Order) error {
	if stats.SignalID == 0 || stats.SignalID != order.SignalID {
		return fmt.Errorf("given stats is invalid")
	}

	if order.Profit > stats.Funds {
		return fmt.Errorf("available funds are less than withdraw amount")
	}

	stats.Funds -= order.Profit
	stats.Withdrawals += order.Profit

	order.Code = ""
	order.Name = ""
	order.NumShares = 0
	order.Price = 0

	return nil
}

func executeBuyOrder(signal model.Signal, stats *model.Stats, order *model.Order, holding *model.Holding, tx *sqlx.Tx) error {
	if tx == nil {
		return fmt.Errorf("tx is nil")
	}

	if order.Type == model.BUY && ((float64(order.NumShares) * order.Price) > stats.Funds) {
		return fmt.Errorf("not available funds to buy the order")
	}

	// Add the new holding if it does not exist
	if holding.ID == 0 {
		holding.SignalID = order.SignalID
		holding.Code = order.Code
		holding.Name = order.Name
		holding.NumShares = order.NumShares
		holding.Price = order.Price
		stats.Time = order.Time

		// Insert the holding
		_, err := tx.NamedExec("INSERT INTO holdings (signal_id, code, name, num_shares, price)"+
			" VALUES (:signal_id, :code, :name, :num_shares, :price)", &holding)
		if err != nil {
			return fmt.Errorf("failed to insert holding : %s", err)
		}
	} else {
		holding.Price = (holding.Price*float64(holding.NumShares) + order.Price*float64(order.NumShares)) /
			float64(holding.NumShares+order.NumShares)
		holding.NumShares += order.NumShares

		_, err := tx.NamedExec("UPDATE holdings SET"+
			" price = :price, num_shares = :num_shares WHERE id = :id", &holding)
		if err != nil {
			return fmt.Errorf("failed to update holding : %s", err)
		}
	}

	switch order.Type {
	case model.BUY:
		stats.Funds -= float64(order.NumShares) * order.Price
	case model.ADD:
		stats.Deposits += float64(order.NumShares) * order.Price
	}

	signal.NumTrades++
	if signal.FirstTradeTime == 0 {
		signal.FirstTradeTime = order.Time
	}

	signal.LastTradeTime = order.Time
	_, err := tx.NamedExec("UPDATE signals SET "+
		"num_trades = :num_trades, first_trade_time = :first_trade_time, "+
		"last_trade_time = :last_trade_time WHERE id = :id",
		&signal)
	if err != nil {
		return fmt.Errorf("failed to update signal : %s", err)
	}

	return nil
}

func executeSellOrder(signal model.Signal, stats *model.Stats, order *model.Order, holding *model.Holding, tx *sqlx.Tx) error {
	if tx == nil {
		return fmt.Errorf("tx is nil")
	}

	if holding == nil {
		return fmt.Errorf("%s stock does not exist in the holdings", order.Code)
	}

	if order.NumShares > holding.NumShares {
		return fmt.Errorf("%d %s stock does not exist in the holdings", order.NumShares, order.Code)
	}

	profit := float64(order.NumShares) * (order.Price - holding.Price)

	holding.NumShares -= order.NumShares
	switch holding.NumShares {
	case 0:
		_, err := tx.NamedExec("DELETE FROM holdings "+
			" WHERE id = :id", &holding)
		if err != nil {
			return fmt.Errorf("failed to delete holding : %s", err)
		}
	default:
		_, err := tx.NamedExec("UPDATE holdings SET"+
			" num_shares = :num_shares WHERE id = :id", &holding)
		if err != nil {
			return fmt.Errorf("failed to update holding : %s", err)
		}
	}

	switch order.Type {
	case model.SELL:
		stats.Funds += float64(order.NumShares) * order.Price
	case model.REDUCE:
		stats.Withdrawals += float64(order.NumShares) * order.Price
	}

	signal.NumTrades++
	signal.LastTradeTime = order.Time
	_, err := tx.NamedExec("UPDATE signals SET"+
		" num_trades = :num_trades, last_trade_time = :last_trade_time WHERE id = :id",
		&signal)
	if err != nil {
		return fmt.Errorf("failed to update signal : %s", err)
	}

	order.Profit = profit
	return nil
}

func findHolding(code string, holdings []model.Holding) int {
	for i, h := range holdings {
		if h.Code == code {
			return i
		}
	}
	return -1
}
