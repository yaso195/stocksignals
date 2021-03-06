package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/heroku/stocksignals/model"
	"github.com/heroku/stocksignals/stockapi"
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

func GetPortfolioBySignalID(c *gin.Context) {
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

	signal, err := store.GetSignalByID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if signal == nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("signal with id %s does not exist", id))
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

	if err = computePortfolio(stats, holdings); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	portfolio := model.Portfolio{Stats: *stats, Holdings: holdings}
	c.JSON(http.StatusOK, portfolio)
}

func computePortfolio(stats *model.Stats, holdings []model.Holding) error {
	if stats == nil {
		return fmt.Errorf("failed to compute portfolio : given stats is nil")
	}

	var stocks []string
	var totalStockEquity float64
	if len(holdings) > 0 {
		for _, holding := range holdings {
			stocks = append(stocks, holding.Code)
		}

		prices, err := stockapi.GetBidPrices(stocks)
		if err != nil {
			return err
		}

		for i, holding := range holdings {
			holdings[i].Gain = 0
			if holding.Price != 0 {
				holdings[i].Gain = (prices[i] - holding.Price) * 100.0 / holding.Price
				holdings[i].Gain = prettifyFloat(holdings[i].Gain)
			}
			totalStockEquity += prices[i] * float64(holding.NumShares)
		}

		stats.Equity = totalStockEquity + stats.Funds
		for i, holding := range holdings {
			holdings[i].Ratio = 0

			if totalStockEquity != 0 {
				holdings[i].Ratio = prices[i] * float64(holding.NumShares) * 100.0 / totalStockEquity
				holdings[i].Ratio = prettifyFloat(holdings[i].Ratio)
			}
		}

		gain := 0.0
		if stats.Balance != 0 {
			gain = (stats.Equity - stats.Balance) * 100.0 / stats.Balance
		}
		gain = prettifyFloat(gain)
		stats.Growth += gain
		stats.Growth = prettifyFloat(stats.Growth)
		stats.Drawdown = gain * -1
	}

	return nil
}

func prettifyFloat(x float64) float64 {
	result := int64((x + 0.005) * 100.0)

	return float64(result) / 100.0
}
