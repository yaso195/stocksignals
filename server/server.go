package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

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
	router.POST("/signals", RegisterSignals)
	router.GET("/signal", GetSignalByID)
	router.DELETE("/signals", DeleteSignalsByID)

	router.GET("/users", GetUsers)
	router.POST("/user", RegisterUser)
	router.GET("/user/:email", GetUserByEmail)

	router.GET("/orders", GetOrdersBySignalID)
	router.POST("/orders", RegisterOrders)
	router.DELETE("/orders", DeleteOrdersByID)

	router.GET("/holdings", GetHoldingsBySignalID)

	router.GET("/stats", GetLatestStatsBySignalID)
	router.GET("/stats_all", GetAllStatsBySignalID)
	router.POST("/stats_save", SaveSignalStats)

	router.GET("/portfolio", GetPortfolioBySignalID)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		router.Run(":" + port)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		saveStats(port)
	}()

	wg.Wait()
}

func saveStats(port string) {
	for {
		http.Post(fmt.Sprintf("http://127.0.0.1:%s/stats_save", port), "", nil)
		time.Sleep(6 * time.Hour)
	}
}
