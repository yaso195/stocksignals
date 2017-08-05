package store

import (
	"database/sql"
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

func RegisterOrders(orders []model.Order) error {
	tx := db.MustBegin()

	// Map the orders based on their signal ID
	signalToOrdersMap := make(map[int][]model.Order)
	for _, order := range orders {
		if _, ok := signalToOrdersMap[order.SignalID]; !ok {
			signalToOrdersMap[order.SignalID] = []model.Order{}
		}

		signalToOrdersMap[order.SignalID] = append(signalToOrdersMap[order.SignalID], order)
	}

	var err error
	for signalID, orders := range signalToOrdersMap {
		signal, err := GetSignalByID(signalID)
		if err != nil {
			return err
		}

		stats, err := GetLatestStats(signalID)
		if err != nil {
			return err
		}

		holdings, err := GetHoldingsBySignalID(signalID, "", true)
		if err != nil {
			return err
		}

		for _, order := range orders {
			if err = registerOrder(signal, &order, stats, &holdings, tx); err != nil {
				return err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete order registration : %s", err)
	}
	return nil
}

func registerOrder(signal *model.Signal, order *model.Order, stats *model.Stats, holdings *[]model.Holding, tx *sqlx.Tx) error {
	if order == nil {
		return fmt.Errorf("given order is nil")
	}

	var statsProfit float64
	var err error
	switch order.Type {
	case model.DEPOSIT:
		err = executeDepositOrder(stats, order)
	case model.WITHDRAW:
		err = executeWithdrawOrder(stats, order)
	case model.BUY, model.ADD:
		loc := findHolding(order.Code, *holdings)
		var holding *model.Holding
		if loc == -1 {
			holding = &model.Holding{}
		} else {
			holding = &(*holdings)[loc]
		}

		err = executeBuyOrder(*signal, stats, order, holding, tx)

		if loc == -1 {
			*holdings = append(*holdings, *holding)
		} else {
			(*holdings)[loc] = *holding
		}
	case model.SELL, model.REDUCE:
		loc := findHolding(order.Code, *holdings)
		var holding *model.Holding
		if loc != -1 {
			holding = &(*holdings)[loc]
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

	stats.Time = order.Time

	if err = insertStats(tx, stats, statsProfit, *holdings, order.PastOrder); err != nil {
		return err
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

	if order.Time > signal.LastTradeTime {
		signal.LastTradeTime = order.Time
	}
	if order.Time < signal.FirstTradeTime {
		signal.FirstTradeTime = order.Time
	}
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
	if order.Time > signal.LastTradeTime {
		signal.LastTradeTime = order.Time
	}
	if order.Time < signal.FirstTradeTime {
		signal.FirstTradeTime = order.Time
	}
	_, err := tx.NamedExec("UPDATE signals SET"+
		" num_trades = :num_trades, first_trade_time = :first_trade_time, last_trade_time = :last_trade_time WHERE id = :id",
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

func delete(stats *model.Stats, order *model.Order) error {
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

func deleteOrdersBySignalID(signal_id int, tx *sqlx.Tx) error {
	if tx == nil {
		return fmt.Errorf("given transaction is nil")
	}

	_, err := tx.Exec(fmt.Sprintf("DELETE FROM orders WHERE signal_id = %d", signal_id))
	if err != nil {
		return fmt.Errorf("failed to delete orders from store : %s", err)
	}

	return nil
}

// DeleteOrdersByID deletes the given orders from the database
// It cleans up only the orders, so be careful when you are using it
func DeleteOrdersByID(ids []int) error {
	tx := db.MustBegin()

	var err error
	for _, id := range ids {
		_, err = getOrderByID(id)
		if err != nil {
			return err
		}

		_, err = tx.Exec(fmt.Sprintf("DELETE FROM orders WHERE id = %d", id))
		if err != nil {
			return fmt.Errorf("failed to delete order from store : %s", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to complete order deletion : %s", err)
	}

	return nil
}

// Reads the order from the database by ID, returns empty ID if it cannot find it
func getOrderByID(id int) (*model.Order, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	if id < 0 {
		return nil, fmt.Errorf("invalid order id")
	}
	var result model.Order
	err := db.Get(&result, fmt.Sprintf("SELECT * FROM order WHERE id=%d", id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order with id %d does not exist.", id)
	}

	if err != nil {
		return nil, fmt.Errorf("error reading order with id %d: %q", id, err)
	}

	return &result, nil
}
