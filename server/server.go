package server

import (
	"bytes"
	//"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	// ErrorMarshalJSONOutput is returned when an error occurs on marshalling a
	// JSONOutput object.
	ErrorMarshalJSONOutput = "Expect marshal [%v] to json but failed: %s "
)

// stocksignals the web server
func WelcomeStockSignals(c *gin.Context) {
	var buffer bytes.Buffer
	buffer.WriteString("Welcome to the StockSignals backend!\n")

	c.String(http.StatusOK, buffer.String())
}

func Run() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.GET("/", WelcomeStockSignals)

	router.GET("/signals", GetSignals)
	router.POST("/signal", RegisterSignal)
	router.GET("/signal/:id", GetSignalByID)

	router.GET("/users", GetUsers)
	router.POST("/user", RegisterUser)
	router.GET("/user/:email", GetUserByEmail)

	router.GET("/orders", GetOrdersBySignalID)
	router.POST("/order", RegisterOrder)

	router.GET("/holdings", GetHoldingsBySignalID)
	router.GET("/stats", GetLatestStatsBySignalID)
	router.GET("/stats_all", GetAllStatsBySignalID)

	router.GET("/portfolio", GetPortfolioBySignalID)

	router.Run(":" + port)
}
