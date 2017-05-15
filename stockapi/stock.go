package stockapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	BASE_STOCK_URI = "http://finance.yahoo.com/d/quotes.csv?s=%s&f=%s"
)

// GetAskPrices queries the yahoo api and returns the current ask price for that stocks.
func GetAskPrices(stocks []string) ([]float64, error) {
	pricesStr, err := getStocksResponses(stocks, "a")
	if err != nil {
		return nil, err
	}

	var prices []float64
	for _, priceStr := range pricesStr {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, err
		}

		prices = append(prices, price)
	}
	return prices, nil
}

// GetBidPrices queries the yahoo api and returns the current bid price for the given stocks.
func GetBidPrices(stocks []string) ([]float64, error) {
	pricesStr, err := getStocksResponses(stocks, "b")
	if err != nil {
		return nil, err
	}

	var prices []float64
	for _, priceStr := range pricesStr {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, err
		}

		prices = append(prices, price)
	}
	return prices, nil
}

// GetName queries the yahoo api and returns the name for that stocks.
func GetNames(stocks []string) ([]string, error) {
	names, err := getStocksResponses(stocks, "n")
	if err != nil {
		return nil, err
	}

	return names, nil
}

func getStocksResponses(stocks []string, option string) ([]string, error) {
	stocksStr := ""
	if len(stocks) <= 0 {
		return nil, fmt.Errorf("length of codes cannot be less than 1")
	}

	for i, stock := range stocks {
		if i == 0 {
			stocksStr = stock
			continue
		}
		stocksStr += "," + stock
	}

	response, err := http.Get(fmt.Sprintf(BASE_STOCK_URI, stocksStr, option))
	if err != nil {
		return nil, fmt.Errorf("failed to get stock respnse for option %s : %s", option, err)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read stock response data : %s", err)
	}

	linesStr := string(responseData)

	lines := strings.Split(strings.Trim(linesStr, "\n"), "\n")

	var array []string
	for i, line := range lines {
		line = strings.Trim(line, "\"")
		if line == "N/A" {
			return nil, fmt.Errorf("failed to get stock info for %s option %s", stocks[i], option)
		}
		array = append(array, line)
	}
	return array, nil
}
