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

func RegisterOrder(order model.Order) error {
	stats, err := GetStats(order.SignalID)
	if err != nil {
		return err
	}

	holding, err := getHolding(order.SignalID, order.Code)
	if err != nil {
		return err
	}

	tx := db.MustBegin()

	switch order.Type {
	case model.DEPOSIT:
		err = prepareDepositOrder(stats, &order, tx)
	case model.WITHDRAW:
		err = prepareWithdrawOrder(stats, &order, tx)
	case model.BUY:
		err = prepareBuyOrder(stats, &order, holding, tx)
	case model.SELL:
		err = prepareSellOrder(stats, &order, holding, tx)
	}

	if err != nil {
		return fmt.Errorf("failed to prepare %s order : %s", order.Type, err)
	}

	_, err = tx.NamedExec("INSERT INTO orders (signal_id, order_time, type, code, name, num_shares, price, profit)"+
		" VALUES (:signal_id, :order_time, :type, :code, :name, :num_shares, :price, :profit)", &order)
	if err != nil {
		return fmt.Errorf("failed to insert order : %s", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete order registration : %s", err)
	}
	return nil
}

func prepareDepositOrder(stats model.Stats, order *model.Order, tx *sqlx.Tx) error {
	if tx == nil {
		return fmt.Errorf("tx is nil")
	}

	if stats.ID == 0 {
		if err := createStats(order.SignalID); err != nil {
			return err
		}
	}

	stats.ID = order.SignalID
	if err := updateStatsFund(tx, order.Profit, false, stats); err != nil {
		return err
	}

	order.Code = ""
	order.Name = ""
	order.NumShares = 0
	order.Price = 0
	return nil
}

func prepareWithdrawOrder(stats model.Stats, order *model.Order, tx *sqlx.Tx) error {
	if tx == nil {
		return fmt.Errorf("tx is nil")
	}

	if stats.ID == 0 {
		return fmt.Errorf("failed to withdraw, because signal has no funds")
	}

	if err := updateStatsFund(tx, order.Profit*(-1), false, stats); err != nil {
		return err
	}

	order.Code = ""
	order.Name = ""
	order.NumShares = 0
	order.Price = 0

	return nil
}

func prepareBuyOrder(stats model.Stats, order *model.Order, holding model.Holding, tx *sqlx.Tx) error {
	if tx == nil {
		return fmt.Errorf("tx is nil")
	}

	if (float64(order.NumShares) * order.Price) > stats.Funds {
		return fmt.Errorf("not available funds to buy the order")
	}

	// Add the new holding if it does not exist
	if holding.ID == 0 {
		holding.SignalID = order.SignalID
		holding.Code = order.Code
		holding.Name = order.Name
		holding.NumShares = order.NumShares
		holding.Price = order.Price

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

	stats.Funds -= float64(order.NumShares) * order.Price
	stats.NumTrades++
	_, err := tx.NamedExec("UPDATE stats SET"+
		" funds = :funds, num_trades = :num_trades WHERE id = :id", &stats)
	if err != nil {
		return fmt.Errorf("failed to update stats : %s", err)
	}
	return nil
}

func prepareSellOrder(stats model.Stats, order *model.Order, holding model.Holding, tx *sqlx.Tx) error {
	if tx == nil {
		return fmt.Errorf("tx is nil")
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

	stats.Profit += profit
	stats.NumTrades++
	stats.Funds += float64(order.NumShares) * order.Price
	_, err := tx.NamedExec("UPDATE stats SET"+
		" profit = :profit, funds = :funds, num_trades = :num_trades WHERE id = :id", &stats)
	if err != nil {
		return fmt.Errorf("failed to update stats : %s", err)
	}

	order.Profit = profit
	return nil
}
