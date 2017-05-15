package server

import (
	//"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	//"github.com/heroku/stocksignals/model"
	//"github.com/heroku/stocksignals/stockapi"
	"github.com/heroku/stocksignals/store"
)

// GetHoldingsBySignalID retrieves the holdings by signal ID parameter
func GetHoldingsBySignalID(c *gin.Context) {
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

	holding, err := store.GetHoldingsBySignalID(id, field, order)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, holding)
}

/*func GetPortfolioBySignalID(c *gin.Context) {
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

	holdings, err := store.GetHoldingsBySignalID(id, "", true)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stats, err := store.GetLatestStats(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	portfolio, err := computePortfolio(stats, holdings)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, portfolio)
} */

/*func computePortfolio(stats model.Stats, holdings []model.Holding) (*model.Portfolio, error) {
	var portfolio model.Portfolio

	var stocks []string
	for _, holding := range holdings {
		stocks = append(stocks, holding.Code)
	}

	prices, err := stockapi.GetBidPrices(stocks)
	if err != nil {
		return nil, err
	}

	var totalStockBalance, totalStockEquity float64
	for i, holding := range holdings {
		totalStockBalance += holding.Price * float64(holding.NumShares)
		totalStockEquity += prices[i] * float64(holding.NumShares)
	}

	portfolio.Balance = totalStockBalance + stats.Funds
	portfolio.Equity = totalStockEquity + stats.Funds

	return &portfolio, nil

}*/
