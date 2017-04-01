package server

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/heroku/stocksignals/store"
)

// stocksignals the web server
func WelcomeStockSignals(c *gin.Context) {
	var buffer bytes.Buffer
	buffer.WriteString("Welcome to the StockSignals backend!\n")

	c.String(http.StatusOK, buffer.String())
}

// stocksignals the web server
func GetSignals(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
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

	router.Run(":" + port)
}
