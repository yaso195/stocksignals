package server

import (
	//"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/heroku/stocksignals/store"
)

// GetLatestStatsBySignalID retrieves the latest stats by signal ID parameter
func GetLatestStatsBySignalID(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	idStr := c.Query("signal_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stats, err := store.GetLatestStats(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAllStatsBySignalID retrieves the latest stats by signal ID parameter
func GetAllStatsBySignalID(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	idStr := c.Query("signal_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stats, err := store.GetAllStats(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, stats)
}
