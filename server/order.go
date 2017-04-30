package server

import (
	//"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/heroku/stocksignals/model"
	"github.com/heroku/stocksignals/store"
)

// GetOrdersBySignalID retrieves the orders by signal ID parameter
func GetOrdersBySignalID(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

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
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var order model.Order
	if err := c.BindJSON(&order); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	order.Price = 0.0
	order.Profit = 0.0
	if order.Time.IsZero() {
		order.Time = time.Now()
	}

	if err := store.RegisterOrder(order); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "order is registered"})
}
