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

	name, err := stockapi.GetName(order.Code)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	order.Name = name

	var price float64
	switch order.Type {
	case model.BUY:
		price, err = stockapi.GetAskPrice(order.Code)
	case model.SELL:
		price, err = stockapi.GetBidPrice(order.Code)
	}

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	order.Price = price

	if order.Time == 0 {
		order.Time = time.Now().Unix()
	}

	if err := store.RegisterOrder(order); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "order is registered"})
}
