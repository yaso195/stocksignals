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

// GetAskPrice queries the yahoo api and returns the current ask price for that stock.
func GetAskPrice(code string) (float64, error) {
	priceStr, err := getStockResponse(code, "a")
	if err != nil {
		return 0.0, err
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0.0, err
	}

	return price, nil
}

// GetBidPrice queries the yahoo api and returns the current bid price for that stock.
func GetBidPrice(code string) (float64, error) {
	priceStr, err := getStockResponse(code, "b")
	if err != nil {
		return 0.0, err
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0.0, err
	}

	return price, nil
}

// GetName queries the yahoo api and returns the name for that stock.
func GetName(code string) (string, error) {
	name, err := getStockResponse(code, "n")
	if err != nil {
		return "", err
	}

	return name, nil
}

func getStockResponse(code string, option string) (string, error) {
	response, err := http.Get(fmt.Sprintf(BASE_STOCK_URI, code, option))
	if err != nil {
		return "", fmt.Errorf("failed to get stock price : ", err)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read stock response data : ", err)
	}

	responseStr := strings.Trim(string(responseData), "\"\n")
	if responseStr == "N/A" {
		return "", fmt.Errorf("failed to get stock info for %s option %s", code, option)
	}
	fmt.Println("Response string : ", responseStr)

	return responseStr, nil
}
