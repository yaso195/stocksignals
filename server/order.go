package server

import (
	//"fmt"
	"net/http"
	"strconv"
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

// RegisterOrder registers the given order
func RegisterOrder(c *gin.Context) {
	var err error
	if err = store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	var order model.Order
	if err = c.BindJSON(&order); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	names, err := stockapi.GetNames([]string{order.Code})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
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
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(prices) > 0 {
		order.Price = prices[0]
	}

	if order.Time == 0 {
		order.Time = time.Now().Unix()
	}

	if err := store.RegisterOrder(&order); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "order is registered"})
}
