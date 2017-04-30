package store

import (
	"fmt"

	"github.com/heroku/stocksignals/model"
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
	holding, err := getHolding(order.SignalID, order.Code)
	if err != nil {
		return err
	}

	if order.Type == model.SELL && order.NumShares > holding.NumShares {
		return fmt.Errorf("number of %d stock %s does not exist in the holdings", order.NumShares, order.Code)
	}

	tx := db.MustBegin()

	profit := 0.0
	// Add the new holding if it does not exist
	if holding.Code == "" {
		holding.SignalID = order.SignalID
		holding.Code = order.Code
		holding.Name = order.Name
		holding.NumShares = order.NumShares
		holding.Price = order.Price

		// Insert the holding
		tx.MustExec("INSERT INTO holdings (signal_id, code, name, num_shares, price)"+
			" VALUES (:signal_id, :code, :name, :num_shares, :price)", &holding)
	} else {
		// Sell the holding
		if order.Type == model.SELL {
			profit = float64(holding.NumShares-order.NumShares) * holding.Price

			holding.NumShares -= order.NumShares
			switch holding.NumShares {
			case 0:
				tx.MustExec("DELETE FROM holdings "+
					" WHERE id = :id", &holding)
			default:
				tx.MustExec("UPDATE holdings SET"+
					" num_shares = :num_shares WHERE id = :id", &holding)
			}
			// Buy more holding
		} else if order.Type == model.BUY {
			holding.Price = (holding.Price*float64(holding.NumShares) + order.Price*float64(order.NumShares)) /
				float64(holding.NumShares+order.NumShares)
			holding.NumShares += order.NumShares

			tx.MustExec("UPDATE holdings SET"+
				" price = :price WHERE id = :id", &holding)
		}
	}
	order.Profit = profit

	tx.MustExec("INSERT INTO orders (signal_id, order_time, type, code, name, num_shares, price, profit)"+
		" VALUES (:signal_id, :order_time, :type, :code, :name, :num_shares, :price, :profit)", &order)

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete order registration : %s", err)
	}
	return nil
}
