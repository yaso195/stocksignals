package server

import (
	//"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/heroku/stocksignals/store"
)

// GetHoldingsBySignalID retrieves the holdings by signal ID parameter
func GetHoldingsBySignalID(c *gin.Context) {
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

	orders, err := store.GetHoldingsBySignalID(id, field, order)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, orders)
}
