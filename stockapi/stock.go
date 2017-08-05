package stockapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	//"strings"

	"github.com/buger/jsonparser"
)

const (
	//BASE_STOCK_URI = "http://finance.yahoo.com/d/quotes.csv?s=%s&f=%s"
	BASE_STOCK_URI = "https://query.yahooapis.com/v1/public/yql?q=select%%20*%%20from%%20csv%%20" +
		"where%%20url%%3D'http%%3A%%2F%%2Fdownload.finance.yahoo.com%%2Fd%%2Fquotes.csv%%3Fs%%3D" +
		"%s" +
		"%%26f%%3Ds" +
		"%s" +
		"%%26e%%3D.csv'%%20and%%20columns%%3D'" +
		"symbol%%2Cvalue" +
		"'&format=json&env=store%%3A%%2F%%2Fdatatables.org%%2Falltableswithkeys"
)

// GetAskPrices queries the yahoo api and returns the current ask price for that stocks.
func GetAskPrices(stocks []string) ([]float64, error) {
	pricesStr, err := getStocksResponses(stocks, "l1")
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
	pricesStr, err := getStocksResponses(stocks, "l1")
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

	count, err := jsonparser.GetInt(responseData, "query", "count")
	if err != nil {
		return nil, fmt.Errorf("failed to read the query count from the response data: %s", err)
	}

	if int(count) != len(stocks) {
		return nil, fmt.Errorf("query count is not equal to the number of stocks")
	}

	var array []string
	var val string
	if len(stocks) == 1 {
		val, err = jsonparser.GetString(responseData, "query", "results", "row", "value")
		if err != nil {
			return nil, fmt.Errorf("failed to get requested %s for %s : %s", option, stocksStr, err)
		}
		array = append(array, val)
	} else {
		jsonparser.ArrayEach(responseData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			val, err = jsonparser.GetString(value, "value")
			array = append(array, val)
		}, "query", "results", "row")
	}

	return array, nil
}
