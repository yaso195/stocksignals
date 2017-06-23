package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/heroku/stocksignals/model"
	"github.com/heroku/stocksignals/store"
)

// GetSignals retrieves the signals from the user
func GetSignals(c *gin.Context) {
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

	signals, err := store.GetSignals(field, order)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, signals)
}

// RegisterSignals register the given signal
func RegisterSignals(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	var signals []model.Signal
	if err := c.BindJSON(&signals); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(signals) == 0 {
		c.String(http.StatusInternalServerError, fmt.Sprintf("no signals is given"))
		return
	}

	if err := store.RegisterSignals(signals); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(signals) == 1 {
		c.JSON(http.StatusOK, gin.H{"status": "signal is registered"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "signals are registered"})
	}
}

// GetSignalByID retrieves the signals by ID parameter
func GetSignalByID(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	signal, err := store.GetSignalByID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, signal)
}

// DeleteSignalsByID deletes the signals (orders, holdings and stats) by ID parameter
func DeleteSignalsByID(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	idsStr := c.Query("id")
	idsStrArr := strings.Split(idsStr, ",")

	if len(idsStrArr) == 0 {
		c.String(http.StatusInternalServerError, fmt.Sprintf("no signal id is given"))
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

	err := store.DeleteSignalsByID(ids)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(idsStrArr) == 1 {
		c.JSON(http.StatusOK, gin.H{"status": "signal is deleted"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "signals are deleted"})
	}
}
