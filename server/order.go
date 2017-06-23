package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/heroku/stocksignals/model"
	"github.com/heroku/stocksignals/stockapi"
	"github.com/heroku/stocksignals/store"
)

// GetOrdersBySignalID retrieves the orders by signal ID parameter
func GetOrdersBySignalID(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	field := c.DefaultQuery("field", "")
	orderStr := c.DefaultQuery("order", "true")
	order, err := strconv.ParseBool(orderStr)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	idStr := c.Query("signal_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	orders, err := store.GetOrdersBySignalID(id, field, order)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, orders)
}

// RegisterOrders registers the given orders
func RegisterOrders(c *gin.Context) {
	var err error
	if err = store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	var orders []model.Order
	if err = c.BindJSON(&orders); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(orders) == 0 {
		c.String(http.StatusInternalServerError, fmt.Sprintf("no order is given to register"))
		return
	}

	preparedOrders, err := prepareOrders(orders)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := store.RegisterOrders(preparedOrders); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(orders) == 1 {
		c.JSON(http.StatusOK, gin.H{"status": "order is registered"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "orders are registered"})
	}
}

func prepareOrders(orders []model.Order) ([]model.Order, error) {
	var list []model.Order
	for _, order := range orders {
		names, err := stockapi.GetNames([]string{order.Code})
		if err != nil {
			return nil, err
		}
		order.Name = names[0]

		var prices []float64
		switch order.Type {
		case model.BUY, model.ADD:
			prices, err = stockapi.GetAskPrices([]string{order.Code})
		case model.SELL, model.REDUCE:
			prices, err = stockapi.GetBidPrices([]string{order.Code})
		}

		if err != nil {
			return nil, err
		}

		if order.Price == 0 && len(prices) > 0 {
			order.Price = prices[0]
		}

		if order.Time == 0 {
			order.Time = time.Now().Unix()
			order.PastOrder = false
		} else {
			order.PastOrder = true
		}

		list = append(list, order)
	}

	return list, nil
}

// DeleteOrdersByID deletes the orders by ID parameter. Note that
// it does not clean up the stats, holdings related with this orders.
func DeleteOrdersByID(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	idsStr := c.Query("id")
	idsStrArr := strings.Split(idsStr, ",")

	if len(idsStrArr) == 0 {
		c.String(http.StatusInternalServerError, fmt.Sprintf("no order id is given"))
		return
	}

	var ids []int
	for _, idStr := range idsStrArr {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		ids = append(ids, id)
	}

	err := store.DeleteOrdersByID(ids)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(idsStrArr) == 1 {
		c.JSON(http.StatusOK, gin.H{"status": "order is deleted"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "orders are deleted"})
	}
}
