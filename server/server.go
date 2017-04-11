package server

import (
	"bytes"
	//"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/heroku/stocksignals/model"
	"github.com/heroku/stocksignals/store"
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

// GetSignals retrieves the signals from the user
func GetSignals(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	field := c.DefaultQuery("field", "")
	orderStr := c.DefaultQuery("order", "true")
	order, err := strconv.ParseBool(orderStr)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	signals, err := store.GetSignals(field, order)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, signals)
}

// RegisterSignal register the given signal
func RegisterSignal(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var signal model.Signal
	if err := c.BindJSON(&signal); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := store.RegisterSignal(signal); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "signal is registered"})
}

// GetSignalByID retrieves the signals by ID parameter
func GetSignalByID(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	idStr := c.Param("id")
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

	c.JSON(http.StatusOK, signal)
}

// GetSignals retrieves the signals from the user
func GetUsers(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	users, err := store.GetUsers()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, users)
}

// RegisterUser registers the given user on the database
func RegisterUser(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var user model.User
	if err := c.BindJSON(&user); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := store.RegisterUser(user); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "user is registered"})
}

// GetUserByEmail gets the given user info from the database via email
func GetUserByEmail(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	email := c.Param("email")
	user, err := store.GetUser(email)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
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

	router.Run(":" + port)
}
