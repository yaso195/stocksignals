package server

import (
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

func SaveSignalStats(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	signals, err := store.GetSignals("", true)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	errFound := false
	for _, signal := range signals {
		err = store.SaveStats(signal.ID)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			errFound = true
		}
	}

	if !errFound {
		c.JSON(http.StatusOK, gin.H{"status": "signals stats are saved successfully"})
	}
}
